package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"crm-prospect-simulator/backend/internal/admin/model"
	"crm-prospect-simulator/backend/internal/admin/repository"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrForbidden                 = errors.New("admin operation forbidden")
	ErrValidation                = errors.New("validation failed")
	ErrSelfDeactivate            = errors.New("cannot deactivate yourself")
	ErrInvalidManager            = errors.New("invalid manager")
	ErrLastAdmin                 = errors.New("cannot deactivate last active administrator")
	ErrInvalidResetMode          = errors.New("invalid reset mode")
	ErrTemporaryPasswordRequired = errors.New("temporary password required")
	ErrWeakTemporaryPassword     = errors.New("weak temporary password")
	ErrInvalidOrganizationalRole = fmt.Errorf("%w: invalid organizational role", ErrValidation)
	ErrProtectedSuperAdmin       = errors.New("protected super admin account")
)

type Service struct {
	repo repository.Repository
}

type profileRepository interface {
	UpdateUserProfile(context.Context, uuid.UUID, model.ProfileUpdateInput, uuid.UUID) error
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

type Actor struct {
	UserID uuid.UUID
	Role   authmodel.Role
}

func (s *Service) ListUsers(ctx context.Context, actor Actor, filter model.ListFilter) (model.UserListResult, error) {
	if !actor.Role.IsAdminRole() {
		return model.UserListResult{}, ErrForbidden
	}
	return s.repo.ListUsers(ctx, filter)
}

func (s *Service) GetUserDetail(ctx context.Context, actor Actor, id uuid.UUID) (model.UserDetail, error) {
	if !actor.Role.IsAdminRole() {
		return model.UserDetail{}, ErrForbidden
	}
	return s.repo.FindUserDetail(ctx, id)
}

func (s *Service) CreateUser(ctx context.Context, actor Actor, input model.CreateUserInput) (model.UserDetail, error) {
	if !actor.Role.IsAdminRole() {
		return model.UserDetail{}, ErrForbidden
	}
	switch effectiveAccountType(input.AccountType) {
	case model.AccountTypeSuperAdmin:
		input.Role = authmodel.RoleSuperAdmin
		input.SalesRoleID = nil
		input.ManagerID = nil
	case model.AccountTypeSalesAccount:
		if input.SalesRoleID == nil {
			return model.UserDetail{}, fmt.Errorf("%w: role is required", ErrInvalidOrganizationalRole)
		}
		role, err := s.repo.FindSalesRole(ctx, *input.SalesRoleID)
		if err != nil {
			return model.UserDetail{}, err
		}
		derivedRole, err := deriveSystemRoleForSalesRole(role)
		if err != nil {
			return model.UserDetail{}, err
		}
		input.Role = derivedRole
	default:
		return model.UserDetail{}, fmt.Errorf("%w: account type must be SUPER_ADMIN or SALES_ACCOUNT", ErrValidation)
	}

	if err := s.validateCreate(ctx, input); err != nil {
		return model.UserDetail{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.TemporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		return model.UserDetail{}, fmt.Errorf("hash password: %w", err)
	}

	userID := uuid.New()
	if err := s.repo.CreateUser(ctx, userID, input, string(passwordHash), actor.UserID); err != nil {
		return model.UserDetail{}, err
	}

	if input.Role != authmodel.RoleSuperAdmin {
		if err := s.repo.SetCurrentSalesAssignment(ctx, userID, input.SalesRoleID, input.ManagerID, actor.UserID); err != nil {
			return model.UserDetail{}, err
		}
	}
	return s.repo.FindUserDetail(ctx, userID)
}

func (s *Service) UpdateUser(ctx context.Context, actor Actor, id uuid.UUID, input model.UpdateUserInput) (model.UserDetail, error) {
	if !actor.Role.IsAdminRole() {
		return model.UserDetail{}, ErrForbidden
	}
	if err := s.ensurePrimarySuperAdminMutable(ctx, actor, id); err != nil {
		return model.UserDetail{}, err
	}

	current, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return model.UserDetail{}, err
	}

	if input.AccountType != nil {
		switch *input.AccountType {
		case model.AccountTypeSuperAdmin:
			role := authmodel.RoleSuperAdmin
			input.Role = &role
			input.SalesRoleID = model.OptionalUUID{Present: true, Value: nil}
			input.ManagerID = model.OptionalUUID{Present: true, Value: nil}
		case model.AccountTypeSalesAccount:
			if !input.SalesRoleID.Present || input.SalesRoleID.Value == nil {
				return model.UserDetail{}, fmt.Errorf("%w: role is required", ErrInvalidOrganizationalRole)
			}
		default:
			return model.UserDetail{}, fmt.Errorf("%w: account type must be SUPER_ADMIN or SALES_ACCOUNT", ErrValidation)
		}
	}

	if current.Role == authmodel.RoleSuperAdmin && input.SalesRoleID.Present {
		input.SalesRoleID = model.OptionalUUID{Present: true, Value: nil}
		input.ManagerID = model.OptionalUUID{Present: true, Value: nil}
		role := authmodel.RoleSuperAdmin
		input.Role = &role
	} else if input.SalesRoleID.Present && input.SalesRoleID.Value != nil && (input.AccountType == nil || *input.AccountType == model.AccountTypeSalesAccount) {
		role, err := s.repo.FindSalesRole(ctx, *input.SalesRoleID.Value)
		if err != nil {
			return model.UserDetail{}, err
		}

		derivedRole, err := deriveSystemRoleForSalesRole(role)
		if err != nil {
			return model.UserDetail{}, err
		}

		input.Role = &derivedRole
	}

	if err := s.validateUpdate(ctx, id, &input); err != nil {
		return model.UserDetail{}, err
	}

	if err := s.repo.UpdateUser(ctx, id, input, actor.UserID); err != nil {
		return model.UserDetail{}, err
	}

	if input.Role != nil && *input.Role == authmodel.RoleSuperAdmin {
		if current.Role != authmodel.RoleSuperAdmin {
			if err := s.repo.SetCurrentSalesAssignment(ctx, id, nil, nil, actor.UserID); err != nil {
				return model.UserDetail{}, err
			}
		}
	} else if input.SalesRoleID.Present || input.ManagerID.Present {
		detail, err := s.repo.FindUserDetail(ctx, id)
		if err != nil {
			return model.UserDetail{}, err
		}
		var salesRoleID *uuid.UUID
		if detail.OrganizationalRole != nil {
			salesRoleID = &detail.OrganizationalRole.ID
		}
		if input.SalesRoleID.Present {
			salesRoleID = input.SalesRoleID.Value
		}
		parentID := detail.ReportsToUserID
		if input.ManagerID.Present {
			parentID = input.ManagerID.Value
		}
		if err := s.repo.SetCurrentSalesAssignment(ctx, id, salesRoleID, parentID, actor.UserID); err != nil {
			return model.UserDetail{}, err
		}
	}
	return s.repo.FindUserDetail(ctx, id)
}

func (s *Service) UpdateUserProfile(ctx context.Context, actor Actor, id uuid.UUID, input model.ProfileUpdateInput) (model.UserDetail, error) {
	if !actor.Role.IsAdminRole() { return model.UserDetail{}, ErrForbidden }
	if input.DateOfBirth != nil { if date, err := time.Parse("2006-01-02", *input.DateOfBirth); err != nil { return model.UserDetail{}, fmt.Errorf("%w: date of birth is invalid", ErrValidation) } else if date.After(time.Now()) { return model.UserDetail{}, fmt.Errorf("%w: date of birth cannot be in the future", ErrValidation) } }
	if input.JoinDate != nil { if _, err := time.Parse("2006-01-02", *input.JoinDate); err != nil { return model.UserDetail{}, fmt.Errorf("%w: join date is invalid", ErrValidation) } }
	if input.Phones != nil {
		for _, phone := range *input.Phones { if strings.TrimSpace(phone.PhoneNumber) == "" { continue }; if len(strings.TrimSpace(phone.PhoneNumber)) < 5 { return model.UserDetail{}, fmt.Errorf("%w: phone number is invalid", ErrValidation) } }
	}
	repo, ok := s.repo.(profileRepository)
	if !ok { return model.UserDetail{}, fmt.Errorf("%w: profile updates are unavailable", ErrValidation) }
	if err := repo.UpdateUserProfile(ctx, id, input, actor.UserID); err != nil { return model.UserDetail{}, err }
	return s.repo.FindUserDetail(ctx, id)
}

func effectiveAccountType(accountType model.AccountType) model.AccountType {
	if accountType == "" {
		return model.AccountTypeSalesAccount
	}
	return accountType
}

func (s *Service) UpdateStatus(ctx context.Context, actor Actor, id uuid.UUID, status authmodel.UserStatus) (model.UserDetail, error) {
	if !actor.Role.IsAdminRole() {
		return model.UserDetail{}, ErrForbidden
	}
	if err := s.ensurePrimarySuperAdminMutable(ctx, actor, id); err != nil {
		return model.UserDetail{}, err
	}
	if id == actor.UserID {
		return model.UserDetail{}, ErrSelfDeactivate
	}
	if status == authmodel.UserInactive {
		count, err := s.repo.CountActiveAdministrators(ctx)
		if err != nil {
			return model.UserDetail{}, err
		}
		target, err := s.repo.FindUserByID(ctx, id)
		if err != nil {
			return model.UserDetail{}, err
		}
		if target.Role.IsAdminRole() && count <= 1 {
			return model.UserDetail{}, ErrLastAdmin
		}
	}
	if err := s.repo.UpdateStatus(ctx, id, status, actor.UserID); err != nil {
		return model.UserDetail{}, err
	}
	return s.repo.FindUserDetail(ctx, id)
}

func (s *Service) DeleteUser(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.Role.IsAdminRole() {
		return ErrForbidden
	}
	if id == actor.UserID {
		return ErrSelfDeactivate
	}
	if err := s.ensurePrimarySuperAdminMutable(ctx, actor, id); err != nil {
		return err
	}
	target, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return err
	}
	if target.Role.IsAdminRole() {
		count, err := s.repo.CountActiveAdministrators(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrLastAdmin
		}
	}
	return s.repo.DeleteUser(ctx, id)
}

func (s *Service) ensurePrimarySuperAdminMutable(ctx context.Context, actor Actor, targetID uuid.UUID) error {
	if actor.Role == authmodel.RoleSuperAdmin {
		return nil
	}
	target, err := s.repo.FindUserByID(ctx, targetID)
	if err != nil {
		return err
	}
	if target.Role == authmodel.RoleSuperAdmin {
		return ErrProtectedSuperAdmin
	}
	return nil
}

func (s *Service) ListManagers(ctx context.Context, actor Actor) ([]model.ManagerOption, error) {
	if !actor.Role.IsAdminRole() {
		return nil, ErrForbidden
	}
	return s.repo.ListActiveManagers(ctx)
}

func (s *Service) ResetPassword(ctx context.Context, actor Actor, targetID uuid.UUID, input model.ResetPasswordInput) (model.ResetPasswordResult, error) {
	if !actor.Role.IsAdminRole() {
		return model.ResetPasswordResult{}, ErrForbidden
	}

	mode := model.ResetPasswordMode(strings.ToUpper(strings.TrimSpace(string(input.Mode))))
	var temporaryPassword string
	switch mode {
	case model.ResetPasswordModeAuto:
		generated, err := generateTemporaryPassword()
		if err != nil {
			return model.ResetPasswordResult{}, fmt.Errorf("generate temporary password: %w", err)
		}
		temporaryPassword = generated
	case model.ResetPasswordModeManual:
		if input.TemporaryPassword == "" {
			return model.ResetPasswordResult{}, fmt.Errorf("%w: temporary password is required for MANUAL mode", ErrTemporaryPasswordRequired)
		}
		if err := validateTemporaryPassword(input.TemporaryPassword); err != nil {
			return model.ResetPasswordResult{}, err
		}
		temporaryPassword = input.TemporaryPassword
	default:
		return model.ResetPasswordResult{}, fmt.Errorf("%w: mode must be AUTO or MANUAL", ErrInvalidResetMode)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		return model.ResetPasswordResult{}, fmt.Errorf("hash temporary password: %w", err)
	}

	sessionsRevoked, err := s.repo.ResetPassword(ctx, targetID, actor.UserID, string(passwordHash))
	if err != nil {
		return model.ResetPasswordResult{}, err
	}

	return model.ResetPasswordResult{
		UserID:             targetID,
		TemporaryPassword:  temporaryPassword,
		MustChangePassword: true,
		SessionsRevoked:    sessionsRevoked,
	}, nil
}

func (s *Service) validateCreate(ctx context.Context, input model.CreateUserInput) error {
	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = strings.TrimSpace(input.Email)
	input.EmployeeID = strings.TrimSpace(input.EmployeeID)
	input.Phone = strings.TrimSpace(input.Phone)
	input.TemporaryPassword = strings.TrimSpace(input.TemporaryPassword)

	if input.FullName == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if input.Email == "" {
		return fmt.Errorf("%w: email is required", ErrValidation)
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return fmt.Errorf("%w: email is invalid", ErrValidation)
	}
	if effectiveAccountType(input.AccountType) == model.AccountTypeSalesAccount && input.SalesRoleID == nil {
		return fmt.Errorf("%w: role is required", ErrInvalidOrganizationalRole)
	}
	if input.Role == "" {
		input.Role = authmodel.RoleSalesExecutive
	}
	if !input.Role.Valid() {
		return fmt.Errorf("%w: invalid role", ErrValidation)
	}
	if input.TemporaryPassword == "" {
		return fmt.Errorf("%w: temporary password is required", ErrValidation)
	}
	if len(input.TemporaryPassword) < 8 {
		return fmt.Errorf("%w: temporary password must be at least 8 characters", ErrValidation)
	}

	if input.Role == authmodel.RoleSuperAdmin {
		if input.SalesRoleID != nil || input.ManagerID != nil {
			return fmt.Errorf("%w: super admin requires no organizational role or manager", ErrInvalidOrganizationalRole)
		}
	} else {
		if err := s.validateAccountOrganizationalRole(ctx, input.Role, input.SalesRoleID, input.ManagerID); err != nil {
			return err
		}
	}

	exists, err := s.repo.ExistsByEmail(ctx, input.Email, nil)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w: email already exists", ErrValidation)
	}

	if input.EmployeeID != "" {
		exists, err = s.repo.ExistsByEmployeeID(ctx, input.EmployeeID, nil)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: employee ID already exists", ErrValidation)
		}
	}

	return nil
}

func (s *Service) validateUpdate(ctx context.Context, id uuid.UUID, input *model.UpdateUserInput) error {
	if input.Email != nil {
		email := strings.TrimSpace(*input.Email)
		if email == "" {
			return fmt.Errorf("%w: email cannot be empty", ErrValidation)
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return fmt.Errorf("%w: email is invalid", ErrValidation)
		}
		exists, err := s.repo.ExistsByEmail(ctx, email, &id)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: email already exists", ErrValidation)
		}
	}

	if input.EmployeeID != nil {
		eid := strings.TrimSpace(*input.EmployeeID)
		if eid != "" {
			exists, err := s.repo.ExistsByEmployeeID(ctx, eid, &id)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("%w: employee ID already exists", ErrValidation)
			}
		}
	}

	if input.FullName != nil && strings.TrimSpace(*input.FullName) == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}

	if input.Role != nil || input.ManagerID.Present || input.SalesRoleID.Present {
		current, err := s.repo.FindUserByID(ctx, id)
		if err != nil {
			return err
		}
		effectiveRole := current.Role
		if input.Role != nil {
			effectiveRole = *input.Role
			if !effectiveRole.Valid() {
				return fmt.Errorf("%w: invalid role", ErrValidation)
			}
		}
		effectiveManagerID := current.ManagerID
		if input.ManagerID.Present {
			effectiveManagerID = input.ManagerID.Value
		}

		if !input.SalesRoleID.Present {
			switch effectiveRole {
			case authmodel.RoleSuperAdmin, authmodel.RoleAdministrator, authmodel.RoleSalesManager:
				if input.ManagerID.Present && input.ManagerID.Value != nil {
					return fmt.Errorf("%w: %s cannot have a manager", ErrValidation, effectiveRole)
				}
				if current.ManagerID != nil && input.Role != nil {
					input.ManagerID = model.OptionalUUID{Present: true, Value: nil}
				}
			case authmodel.RoleSalesExecutive:
				if input.ManagerID.Present {
					if input.ManagerID.Value == nil {
						return fmt.Errorf("%w: sales executive must have a manager", ErrValidation)
					}
					if err := s.validateManagerActive(ctx, *input.ManagerID.Value); err != nil {
						return err
					}
				} else if input.Role != nil {
					if current.ManagerID == nil {
						return fmt.Errorf("%w: sales executive must have a manager", ErrValidation)
					}
					if err := s.validateManagerActive(ctx, *current.ManagerID); err != nil {
						return err
					}
				}
			}
		} else {
			if effectiveRole == authmodel.RoleSuperAdmin {
				input.SalesRoleID = model.OptionalUUID{Present: true, Value: nil}
				input.ManagerID = model.OptionalUUID{Present: true, Value: nil}
				return nil
			}
			if err := s.validateAccountOrganizationalRole(
				ctx,
				effectiveRole,
				input.SalesRoleID.Value,
				effectiveManagerID,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) validateAccountOrganizationalRole(ctx context.Context, systemRole authmodel.Role, salesRoleID *uuid.UUID, managerID *uuid.UUID) error {
	if salesRoleID == nil {
		return fmt.Errorf("%w: role is required", ErrInvalidOrganizationalRole)
	}

	role, err := s.repo.FindSalesRole(ctx, *salesRoleID)
	if err != nil {
		return err
	}
	if !role.IsActive {
		return fmt.Errorf("%w: organizational role is inactive", ErrInvalidOrganizationalRole)
	}
	if role.Level == 1 {
		return fmt.Errorf("%w: level 1 is reserved for SUPER_ADMIN", ErrInvalidOrganizationalRole)
	}

	derivedRole, err := deriveSystemRoleForSalesRole(role)
	if err != nil {
		return err
	}
	if systemRole != derivedRole {
		return fmt.Errorf(
			"%w: account role does not match organizational role",
			ErrInvalidOrganizationalRole,
		)
	}
	if managerID == nil {
		return fmt.Errorf("%w: reports to is required", ErrInvalidOrganizationalRole)
	}
	_, parentRole, err := s.repo.FindEffectiveSalesAssignment(ctx, *managerID, time.Now())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: reports to user has no active hierarchy assignment", ErrInvalidOrganizationalRole)
		}
		return err
	}
	if parentRole.Level != role.Level-1 {
		return fmt.Errorf("%w: parent must be hierarchy level %d", ErrInvalidOrganizationalRole, role.Level-1)
	}

	return nil
}

func deriveSystemRoleForSalesRole(role model.SalesRole) (authmodel.Role, error) {
	if !role.IsActive {
		return "", fmt.Errorf(
			"%w: organizational role is inactive",
			ErrInvalidOrganizationalRole,
		)
	}

	switch role.Level {
	case 1:
		return "", fmt.Errorf(
			"%w: level 1 is reserved for SUPER_ADMIN",
			ErrInvalidOrganizationalRole,
		)
	case 2, 3:
		return authmodel.RoleSalesManager, nil
	case 4:
		return authmodel.RoleSalesExecutive, nil
	default:
		return "", fmt.Errorf(
			"%w: role level must be between 1 and 4",
			ErrInvalidOrganizationalRole,
		)
	}
}

func (s *Service) validateManagerRule(role authmodel.Role, managerID *uuid.UUID) error {
	switch role {
	case authmodel.RoleSuperAdmin, authmodel.RoleAdministrator, authmodel.RoleSalesManager:
		if managerID != nil {
			return fmt.Errorf("%w: %s cannot have a manager", ErrValidation, role)
		}
	case authmodel.RoleSalesExecutive:
		if managerID == nil {
			return fmt.Errorf("%w: sales executive must have a manager", ErrValidation)
		}
	}
	return nil
}

func (s *Service) validateManagerActive(ctx context.Context, managerID uuid.UUID) error {
	manager, err := s.repo.FindManagerByID(ctx, managerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: manager not found", ErrValidation)
		}
		return err
	}
	if manager.Role != authmodel.RoleSalesManager {
		return fmt.Errorf("%w: manager must be a SALES_MANAGER", ErrValidation)
	}
	if manager.Status != authmodel.UserActive {
		return fmt.Errorf("%w: manager is not active", ErrValidation)
	}
	return nil
}

func (s *Service) ensureManagerActive(ctx context.Context, managerID *uuid.UUID) error {
	if managerID == nil {
		return nil
	}
	return s.validateManagerActive(ctx, *managerID)
}

// generateTemporaryPassword builds a secure 12-character password from
// crypto/rand, guaranteeing at least one character from each group
// (upper, lower, digit, symbol) and shuffling with crypto/rand.
func generateTemporaryPassword() (string, error) {
	const (
		upperChars  = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		lowerChars  = "abcdefghijkmnopqrstuvwxyz"
		digitChars  = "23456789"
		symbolChars = "!@#$%&*"
		fillLength  = 8
	)

	groups := []string{upperChars, lowerChars, digitChars, symbolChars}
	bytes := make([]byte, 0, len(groups)+fillLength)

	for _, group := range groups {
		b, err := randomFromSet(group)
		if err != nil {
			return "", err
		}
		bytes = append(bytes, b)
	}

	for i := 0; i < fillLength; i++ {
		b, err := randomFromSet(strings.Join(groups, ""))
		if err != nil {
			return "", err
		}
		bytes = append(bytes, b)
	}

	if err := shuffleBytes(bytes); err != nil {
		return "", err
	}
	return string(bytes), nil
}

func randomFromSet(set string) (byte, error) {
	max := big.NewInt(int64(len(set)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return set[n.Int64()], nil
}

func shuffleBytes(bs []byte) error {
	for i := len(bs) - 1; i > 0; i-- {
		max := big.NewInt(int64(i + 1))
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return err
		}
		j := int(n.Int64())
		bs[i], bs[j] = bs[j], bs[i]
	}
	return nil
}

func validateTemporaryPassword(password string) error {
	if strings.TrimSpace(password) == "" || utf8.RuneCountInString(password) < 6 {
		return fmt.Errorf("%w: temporary password must be at least 6 characters", ErrWeakTemporaryPassword)
	}
	return nil
}

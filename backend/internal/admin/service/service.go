package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"unicode"
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
	if err := s.validateCreate(ctx, input); err != nil {
		return model.UserDetail{}, err
	}
	if input.SalesRoleID != nil {
		role, err := s.repo.FindSalesRole(ctx, *input.SalesRoleID)
		if err != nil {
			return model.UserDetail{}, err
		}
		input.Role = systemRoleForSalesRole(role.Level)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.TemporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		return model.UserDetail{}, fmt.Errorf("hash password: %w", err)
	}
	userID := uuid.New()
	if err := s.repo.CreateUser(ctx, userID, input, string(passwordHash), actor.UserID); err != nil {
		return model.UserDetail{}, err
	}
	if input.SalesRoleID != nil {
		if err := s.repo.SetCurrentSalesAssignment(ctx, userID, input.SalesRoleID, nil, actor.UserID); err != nil {
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
	if err := s.validateUpdate(ctx, id, &input); err != nil {
		return model.UserDetail{}, err
	}

	// Account role changes are intentionally separated from monthly sales
	// hierarchy assignments. Updating the account role changes its derived
	// internal system role and access source, but it must not rewrite the
	// reporting parent or sales-structure history.
	if input.SalesRoleID.Present && input.SalesRoleID.Value != nil {
		role, err := s.repo.FindSalesRole(ctx, *input.SalesRoleID.Value)
		if err != nil {
			return model.UserDetail{}, err
		}
		if !role.IsActive {
			return model.UserDetail{}, fmt.Errorf(
				"%w: organizational role is inactive",
				ErrInvalidOrganizationalRole,
			)
		}

		current, err := s.repo.FindUserByID(ctx, id)
		if err != nil {
			return model.UserDetail{}, err
		}
		if strings.EqualFold(current.Email, "admin@yummy.test") && current.Role == authmodel.RoleSuperAdmin {
			if role.Level != 1 {
				return model.UserDetail{}, fmt.Errorf(
					"%w: super admin requires a level 1 role and no manager",
					ErrInvalidOrganizationalRole,
				)
			}
			preserved := authmodel.RoleSuperAdmin
			input.Role = &preserved
			input.ManagerID = model.OptionalUUID{Present: true, Value: nil}
		} else if current.Role == authmodel.RoleSuperAdmin || current.Role == authmodel.RoleAdministrator {
			preserved := current.Role
			input.Role = &preserved
		} else {
			derived := systemRoleForSalesRole(role.Level)
			input.Role = &derived
		}
	}

	if err := s.repo.UpdateUser(ctx, id, input, actor.UserID); err != nil {
		return model.UserDetail{}, err
	}

	// Do not call SetCurrentSalesAssignment here. Monthly role/parent changes
	// belong to Sales Structure so existing hierarchy and history stay intact.
	return s.repo.FindUserDetail(ctx, id)
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
	if strings.EqualFold(target.Email, "admin@yummy.test") && target.Role == authmodel.RoleSuperAdmin {
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
	if input.SalesRoleID == nil {
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

	if err := s.ensureManagerActive(ctx, input.ManagerID); err != nil {
		return err
	}
	if err := s.validateAccountOrganizationalRole(ctx, input.Role, input.SalesRoleID, input.ManagerID); err != nil {
		return err
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
			// Account Edit may change the organizational role without forcing a
			// hierarchy-parent update. Only the protected Super Admin rule still
			// needs the current manager value to enforce "Level 1 and no manager".
			var managerIDForValidation *uuid.UUID
			if effectiveRole == authmodel.RoleSuperAdmin {
				managerIDForValidation = effectiveManagerID
				if strings.EqualFold(current.Email, "admin@yummy.test") {
					managerIDForValidation = nil
				}
			}

			if err := s.validateAccountOrganizationalRole(
				ctx,
				effectiveRole,
				input.SalesRoleID.Value,
				managerIDForValidation,
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

	if systemRole == authmodel.RoleSuperAdmin {
		if role.Level != 1 || managerID != nil {
			return fmt.Errorf("%w: super admin requires a level 1 role and no manager", ErrInvalidOrganizationalRole)
		}
	}
	return nil
}

func systemRoleForSalesRole(level int) authmodel.Role {
	if level == 4 {
		return authmodel.RoleSalesExecutive
	}
	return authmodel.RoleSalesManager
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
	if utf8.RuneCountInString(password) < 8 {
		return fmt.Errorf("%w: temporary password must be at least 8 characters", ErrWeakTemporaryPassword)
	}
	hasUpper, hasLower, hasDigit := false, false, false
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("%w: temporary password must contain at least one uppercase letter", ErrWeakTemporaryPassword)
	}
	if !hasLower {
		return fmt.Errorf("%w: temporary password must contain at least one lowercase letter", ErrWeakTemporaryPassword)
	}
	if !hasDigit {
		return fmt.Errorf("%w: temporary password must contain at least one digit", ErrWeakTemporaryPassword)
	}
	return nil
}

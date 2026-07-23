package main

import (
	"context"
	"log"

	"crm-prospect-simulator/backend/config"
	"crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/auth/repository"
	"crm-prospect-simulator/backend/platform/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewPostgresRepository(pool)
	users := []model.User{
		{ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash), FullName: "Yummy Administrator", Role: model.RoleAdministrator, Status: model.UserActive},
		{ID: uuid.New(), Email: "sales@yummy.test", PasswordHash: string(hash), FullName: "Nurdin Pratama", Role: model.RoleSalesExecutive, Status: model.UserActive},
		{ID: uuid.New(), Email: "sales2@yummy.test", PasswordHash: string(hash), FullName: "Alicia Ramadhan", Role: model.RoleSalesExecutive, Status: model.UserActive},
		{ID: uuid.New(), Email: "sales3@yummy.test", PasswordHash: string(hash), FullName: "Rizky Ananda", Role: model.RoleSalesExecutive, Status: model.UserActive},
	}
	for _, user := range users {
		if err := repo.UpsertSeed(ctx, user); err != nil {
			log.Fatal(err)
		}
		log.Printf("seeded account %s", user.Role)
	}
	// Historical simulator records use these immutable IDs and Google Place IDs.
	// Delete only those records; independently created Prospect Finder data is never matched.
	demoProspectIDs := []uuid.UUID{
		uuid.MustParse("10000000-0000-4000-8000-000000000001"), uuid.MustParse("10000000-0000-4000-8000-000000000002"),
		uuid.MustParse("10000000-0000-4000-8000-000000000003"), uuid.MustParse("10000000-0000-4000-8000-000000000004"),
		uuid.MustParse("10000000-0000-4000-8000-000000000005"), uuid.MustParse("10000000-0000-4000-8000-000000000006"),
		uuid.MustParse("10000000-0000-4000-8000-000000000007"), uuid.MustParse("10000000-0000-4000-8000-000000000008"),
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM customer_sites WHERE source_prospect_id = ANY($1)`, demoProspectIDs); err != nil {
		log.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM prospect_visits WHERE prospect_id = ANY($1)`, demoProspectIDs); err != nil {
		log.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM prospect_status_history WHERE prospect_id = ANY($1)`, demoProspectIDs); err != nil {
		log.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM prospects WHERE id = ANY($1)`, demoProspectIDs); err != nil {
		log.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM parent_companies pc WHERE pc.parent_code = 'PC-000900' AND NOT EXISTS (SELECT 1 FROM customer_sites cs WHERE cs.parent_company_id = pc.id)`); err != nil {
		log.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded local login accounts; removed legacy simulator business records")
}

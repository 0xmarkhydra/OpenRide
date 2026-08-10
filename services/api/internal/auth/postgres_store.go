package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"flashx/services/api/internal/platform/ids"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreateChallenge(ch Challenge) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO otp_challenges (id, phone, role, code_hash, attempts, expires_at, consumed_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, ch.ID, ch.Phone, string(ch.Role), ch.CodeHash, ch.Attempts, ch.ExpiresAt, ch.ConsumedAt, ch.CreatedAt)
	return err
}

func (s *PostgresStore) VerifyChallenge(id string, role Role, codeHash string, now time.Time, maxAttempts int) (Challenge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ch Challenge
	var roleText string
	var matched bool
	err := s.pool.QueryRow(ctx, `
		UPDATE otp_challenges
		SET attempts = attempts + 1,
		    consumed_at = CASE WHEN code_hash = $3 THEN $4 ELSE consumed_at END
		WHERE id = $1 AND role = $2
		  AND consumed_at IS NULL
		  AND expires_at > $4
		  AND attempts < $5
		RETURNING id, phone, role, code_hash, attempts, expires_at, consumed_at, created_at,
		          code_hash = $3
	`, id, string(role), codeHash, now, maxAttempts).Scan(
		&ch.ID, &ch.Phone, &roleText, &ch.CodeHash, &ch.Attempts,
		&ch.ExpiresAt, &ch.ConsumedAt, &ch.CreatedAt, &matched,
	)
	if err == nil {
		ch.Role = Role(roleText)
		if !matched {
			if ch.Attempts >= maxAttempts {
				return Challenge{}, ErrTooManyAttempts
			}
			return Challenge{}, ErrOTPInvalid
		}
		return ch, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Challenge{}, err
	}

	var attempts int
	var expiresAt time.Time
	var consumedAt *time.Time
	var storedRole string
	err = s.pool.QueryRow(ctx, `
		SELECT role, attempts, expires_at, consumed_at
		FROM otp_challenges WHERE id=$1
	`, id).Scan(&storedRole, &attempts, &expiresAt, &consumedAt)
	if errors.Is(err, pgx.ErrNoRows) || storedRole != string(role) {
		return Challenge{}, ErrChallengeNotFound
	}
	if err != nil {
		return Challenge{}, err
	}
	if consumedAt != nil {
		return Challenge{}, ErrOTPConsumed
	}
	if !expiresAt.After(now) {
		return Challenge{}, ErrOTPExpired
	}
	if attempts >= maxAttempts {
		return Challenge{}, ErrTooManyAttempts
	}
	return Challenge{}, ErrOTPInvalid
}

func (s *PostgresStore) CreateRefreshSession(session RefreshSession) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO refresh_sessions (id, actor_id, role, token_hash, expires_at, revoked_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, session.ID, session.ActorID, string(session.Role), session.TokenHash, session.ExpiresAt, session.RevokedAt, session.CreatedAt)
	return err
}

func (s *PostgresStore) GetRefreshByHash(tokenHash string) (RefreshSession, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var session RefreshSession
	var roleText string
	err := s.pool.QueryRow(ctx, `
		SELECT id, actor_id, role, token_hash, expires_at, revoked_at, created_at
		FROM refresh_sessions WHERE token_hash=$1
	`, tokenHash).Scan(&session.ID, &session.ActorID, &roleText, &session.TokenHash,
		&session.ExpiresAt, &session.RevokedAt, &session.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefreshSession{}, ErrSessionNotFound
	}
	if err != nil {
		return RefreshSession{}, err
	}
	session.Role = Role(roleText)
	return session, nil
}

func (s *PostgresStore) RevokeRefresh(id string, at time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at=COALESCE(revoked_at,$2)
		WHERE id=$1
	`, id, at)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (s *PostgresStore) FindOrCreateActor(ctx context.Context, phone string, role Role) (Actor, error) {
	switch role {
	case RoleRider:
		var actor Actor
		actor.Role = role
		err := s.pool.QueryRow(ctx, `SELECT id, phone FROM users WHERE phone=$1`, phone).Scan(&actor.ID, &actor.Phone)
		if err == nil { return actor, nil }
		if !errors.Is(err, pgx.ErrNoRows) { return Actor{}, err }
		now := time.Now().UTC()
		actor = Actor{ID: ids.New("rider"), Phone: phone, Role: role}
		_, err = s.pool.Exec(ctx, `INSERT INTO users (id,phone,full_name,status,created_at,updated_at) VALUES ($1,$2,'','active',$3,$3) ON CONFLICT (phone) DO NOTHING`, actor.ID, phone, now)
		if err != nil { return Actor{}, err }
		return s.FindOrCreateActor(ctx, phone, role)
	case RoleDriver:
		var actor Actor
		actor.Role = role
		err := s.pool.QueryRow(ctx, `SELECT id, phone FROM drivers WHERE phone=$1`, phone).Scan(&actor.ID, &actor.Phone)
		if err == nil { return actor, nil }
		if !errors.Is(err, pgx.ErrNoRows) { return Actor{}, err }
		now := time.Now().UTC()
		actor = Actor{ID: ids.New("driver"), Phone: phone, Role: role}
		_, err = s.pool.Exec(ctx, `INSERT INTO drivers (id,phone,full_name,service_type,approval_status,availability_status,last_idle_at,version,created_at,updated_at) VALUES ($1,$2,'','bike','pending','offline',$3,1,$3,$3) ON CONFLICT (phone) DO NOTHING`, actor.ID, phone, now)
		if err != nil { return Actor{}, err }
		return s.FindOrCreateActor(ctx, phone, role)
	default:
		return Actor{}, ErrInvalidRole
	}
}

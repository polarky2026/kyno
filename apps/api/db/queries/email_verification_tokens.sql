-- name: CreateEmailVerificationToken :exec
INSERT INTO email_verification_tokens (token_hash, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: DeleteEmailVerificationTokensForUser :exec
DELETE FROM email_verification_tokens
WHERE user_id = $1;

-- name: ConsumeEmailVerificationToken :one
DELETE FROM email_verification_tokens
WHERE token_hash = $1
RETURNING user_id, expires_at;
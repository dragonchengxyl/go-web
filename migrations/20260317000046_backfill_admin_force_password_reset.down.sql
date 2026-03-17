UPDATE users
SET force_password_reset = FALSE,
    updated_at = NOW()
WHERE role = 'admin'
  AND last_login_at IS NULL
  AND force_password_reset = TRUE;

# TOTP (Two-Factor Authentication) Implementation

## Overview

This implementation adds TOTP (Time-based One-Time Password) support to ovpn-admin, providing an additional layer of security alongside password authentication.

## Architecture

### Backend Components

1. **totp.go** - Core TOTP functionality
   - `generateTOTPSecret()` - Generates a new TOTP secret for a user
   - `generateTOTPQRCode()` - Creates a QR code for easy enrollment
   - `verifyTOTP()` - Validates TOTP codes
   - `initTOTPDB()` - Initializes TOTP table in the users database
   - Database operations: save, get, delete, enable, disable TOTP secrets

2. **main.go** - HTTP handlers and integration
   - `userEnableTOTPHandler` - Generates and returns TOTP secret + QR code
   - `userVerifyTOTPHandler` - Verifies TOTP code and enables if first verification
   - `userDisableTOTPHandler` - Removes TOTP for a user
   - `userGetTOTPStatusHandler` - Returns whether TOTP is enabled for a user
   - `userAdminDisableTOTPHandler` - Admin endpoint to disable TOTP for users (emergency access)
   - Database initialization on startup when `--auth.totp` flag is enabled
   - Cleanup of TOTP secrets when users are deleted

### Frontend Components

1. **main.js** - Vue.js logic
   - Setup TOTP action button (shown only when totpAuth module is enabled)
   - Methods for generating, verifying, and disabling TOTP
   - Modal state management for TOTP enrollment

2. **index.html** - UI modal
   - QR code display for scanning with authenticator apps
   - Manual secret entry option
   - TOTP code verification input
   - User-friendly enrollment flow

## Database Schema

TOTP secrets are now stored in the same database as user passwords (users.db):

```sql
CREATE TABLE IF NOT EXISTS totp_secrets (
    username TEXT PRIMARY KEY,
    secret TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 0
);
```

- `username`: Primary key linking to OpenVPN user
- `secret`: Base32-encoded TOTP secret
- `enabled`: 0 = pending verification, 1 = active

**Note:** Previously, TOTP secrets were stored in a separate `totp.db` file. This has been consolidated into `users.db` for easier management and backup.

## API Endpoints

### Enable TOTP
- **Endpoint**: `POST /api/user/totp/enable`
- **Parameters**: `username`
- **Response**: JSON with `secret` and `qrCode` (base64-encoded PNG)

### Verify TOTP
- **Endpoint**: `POST /api/user/totp/verify`
- **Parameters**: `username`, `code`
- **Response**: Success/error message
- **Side Effect**: Enables TOTP on first successful verification

### Disable TOTP
- **Endpoint**: `POST /api/user/totp/disable`
- **Parameters**: `username`
- **Response**: Success/error message

### Get TOTP Status
- **Endpoint**: `POST /api/user/totp/status`
- **Parameters**: `username`
- **Response**: JSON with `enabled` boolean

### Admin Disable TOTP (Emergency Access)
- **Endpoint**: `POST /api/user/totp/admin/disable`
- **Parameters**: `username`
- **Response**: Success/error message
- **Description**: Allows administrators to disable TOTP for users in emergency situations (e.g., lost authenticator device)

## Usage

### Server Configuration

Start ovpn-admin with TOTP enabled:

```bash
./ovpn-admin \
  --auth.totp \
  --auth.db=/path/to/users.db \
  --auth.totp.issuer="My VPN Service"
```

Environment variables:
```bash
export OVPN_AUTH_TOTP=true
export OVPN_AUTH_DB_PATH=/path/to/users.db
export OVPN_TOTP_ISSUER="My VPN Service"
```

**Note:** The `--auth.totp.db` flag has been removed. TOTP secrets are now stored in the same database as user passwords (specified by `--auth.db`).

### User Enrollment Flow

1. User clicks "Setup TOTP" button in the web UI
2. Server generates a unique TOTP secret and QR code
3. User scans QR code with authenticator app (Google Authenticator, Authy, etc.)
4. User enters 6-digit code from app to verify
5. On successful verification, TOTP is enabled for the user

### Admin Emergency Access

If a user loses access to their authenticator device, an administrator can disable TOTP for that user:

1. Admin navigates to the user management interface
2. Admin uses the "Admin Disable TOTP" function for the affected user
3. User can then log in with only their password
4. User can re-enable TOTP by setting it up again

### Compatible Authenticator Apps

- Google Authenticator (iOS, Android)
- Authy (iOS, Android, Desktop)
- Microsoft Authenticator (iOS, Android)
- 1Password
- LastPass Authenticator
- Any TOTP-compatible authenticator

## Security Considerations

1. **Secret Storage**: TOTP secrets are stored in the users database. In production, ensure:
   - Database file permissions are restricted (0600)
   - Database is backed up securely
   - Consider encrypting the database file

2. **QR Code Exposure**: QR codes should only be displayed once during enrollment and should not be logged

3. **Time Synchronization**: TOTP requires accurate system time. Ensure NTP is configured on the server

4. **Two-Step Verification**: This implementation adds a second factor (something you have - the authenticator app) to the first factor (something you know - the password)

5. **Emergency Access**: Admin ability to disable TOTP provides a recovery mechanism while maintaining security

## Integration with OpenVPN

This implementation provides TOTP for the ovpn-admin web interface. To use TOTP with OpenVPN connections:

1. Enable password authentication with `--auth.password`
2. Configure OpenVPN server to use the `openvpn-user` plugin for authentication
3. Modify the authentication script to also verify TOTP codes

Note: Direct OpenVPN TOTP integration would require additional configuration and is outside the scope of this implementation.

## Dependencies

- **github.com/pquerna/otp** (v1.4.0) - TOTP generation and verification
- **github.com/mattn/go-sqlite3** (v1.14.24) - SQLite database driver

Both dependencies have been verified against the GitHub advisory database and have no known vulnerabilities.

## Limitations

- TOTP authentication does not work with `--storage.backend=kubernetes.secrets` (same as password auth)
- No backup codes are provided (users should save their TOTP secret securely)
- No rate limiting on verification attempts (consider adding in production)

## Recent Changes

### Database Consolidation (Latest)
- **Removed**: Separate `totp.db` database file
- **Added**: TOTP secrets now stored in `users.db` alongside user passwords
- **Removed**: `--auth.totp.db` command-line flag
- **Added**: Admin emergency access endpoint to disable TOTP for users
- **Benefit**: Simpler database management, single backup file, no database synchronization issues

## Future Enhancements

Potential improvements for future versions:

1. Backup codes generation and verification
2. Rate limiting on TOTP verification attempts
3. Recovery mechanism if user loses authenticator device
4. Integration with OpenVPN PAM module for connection-time TOTP
5. Support for multiple TOTP devices per user
6. TOTP requirement enforcement (mandatory for all users)

# Test Migration Data

This directory contains SQL migration files for populating the database with test data.

## Test Users

The following test users are created by the migration `20241030000001_test_users.sql`:

### Administrator
- **Email**: `admin@gosign.local`
- **Password**: `admin123`
- **Role**: Admin (3)
- **Account**: Admin Account
- **Email Verified**: Yes

### User 1
- **Email**: `user1@gosign.local`
- **Password**: `user123`
- **Role**: User (1)
- **Account**: User1 Account
- **Email Verified**: Yes

### User 2
- **Email**: `user2@gosign.local`
- **Password**: `user234`
- **Role**: User (1)
- **Account**: User2 Account
- **Email Verified**: Yes

## User Roles

- `1` - User (regular user with basic permissions)
- `2` - Moderator (moderator with extended permissions)
- `3` - Admin (administrator with full access)

## Usage

These test users are automatically created when running migrations in the fixtures directory. They are intended for development and testing purposes only.

**Note**: Each user has their own account and default template folder.


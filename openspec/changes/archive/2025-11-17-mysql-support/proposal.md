# MySQL Database Support

## Change ID
mysql-support

## Title
Add MySQL Database Support with SQLite Compatibility

## Status
implemented

## Authors
ShowTa Development Team

## Created
2025-11-17

## Summary
This change adds MySQL database support to the application while maintaining full compatibility with the existing SQLite implementation. The application will continue to use SQLite by default but will be able to connect to MySQL databases when configured to do so.

## Problem Statement
Currently, the application only supports SQLite as its database backend. While SQLite is suitable for single-user deployments, it has limitations in multi-user, high-concurrency scenarios. Adding MySQL support will:

1. Enable better performance in multi-user environments
2. Provide better scalability for larger deployments
3. Allow users to leverage existing MySQL infrastructure
4. Maintain backward compatibility with existing SQLite installations

## Proposed Solution
1. Add MySQL database driver support using GORM's MySQL dialect
2. Implement database-agnostic connection logic
3. Maintain SQLite as the default database option
4. Update configuration to support database type selection
5. Ensure all models and migrations work with both databases
6. Add proper error handling and connection management for both database types

## Impact
- Enhanced database flexibility and scalability
- Backward compatibility maintained with existing SQLite deployments
- Additional dependency on MySQL driver
- Slightly increased complexity in database initialization
- No breaking changes to existing functionality
# User Management API

A comprehensive API for managing users in the system

**Version:** 2.1.0

## Contact

**Name:** API Support Team
**Email:** support@example.com
**URL:** https://example.com/support

## Servers

- **http://localhost:8080** - Development server
- **https://api.example.com** - Production server
- **https://api.example.com/v2** - Production server
- **https://staging-api.example.com/v2** - Staging server

## Endpoints

### /auth/login

#### POST

**Summary:** User login

**Description:** Authenticate user and return access token

**Responses:**

- `401` - Invalid credentials
- `400` - Invalid request
- `200` - Login successful

### /users

#### GET

**Summary:** List users

**Parameters:**

- `page` (query) - Page number
- `limit` (query) - Items per page

**Responses:**

- `200` - Successful response
- `404` - Resource not found
- `401` - Unauthorized
- `500` - Internal server error

#### POST

**Summary:** Create users

**Responses:**

- `500` - Internal server error
- `201` - Resource created
- `400` - Invalid input
- `401` - Unauthorized

### /users/bulk

#### PATCH

**Summary:** Bulk update users

**Description:** Update multiple users in a single request for efficient batch operations

**Responses:**

- `200` - Users updated successfully
- `400` - Invalid request data
- `422` - Validation errors in request data

### /users/{id}

#### GET

**Summary:** Get users by ID

**Parameters:**

- `id` (path) *required* - Resource ID

**Responses:**

- `404` - Resource not found
- `401` - Unauthorized
- `500` - Internal server error
- `200` - Successful response

#### PUT

**Summary:** Update users

**Parameters:**

- `id` (path) *required* - Resource ID

**Responses:**

- `404` - Resource not found
- `401` - Unauthorized
- `500` - Internal server error
- `200` - Resource updated

#### DELETE

**Summary:** Delete users

**Parameters:**

- `id` (path) *required* - Resource ID

**Responses:**

- `401` - Unauthorized
- `500` - Internal server error
- `204` - Resource deleted
- `404` - Resource not found

### /users/{id}/stats

#### GET

**Summary:** Get user statistics

**Description:** Retrieve detailed statistics and analytics for a specific user

**Parameters:**

- `id` (path) *required* - User ID
- `period` (query) - Time period for statistics (7d, 30d, 90d, 1y, all)
- `include_details` (query) - Include detailed breakdown of statistics

**Responses:**

- `200` - User statistics
- `404` - User not found
- `403` - Forbidden - insufficient permissions

### /auth/me

#### GET

**Summary:** Get current user

**Description:** Get information about the currently authenticated user

**Responses:**

- `200` - User information
- `401` - Unauthorized

### /users/search

#### GET

**Summary:** Advanced user search

**Description:** Search users with advanced filtering, sorting, and pagination

**Parameters:**

- `q` (query) - Search query (searches username, email, first name, last name)
- `page` (query) - Page number (1-based)
- `limit` (query) - Number of items per page (1-100)
- `sort` (query) - Sort field and direction (e.g., 'username:asc', 'created_at:desc')
- `filter` (query) - Filter users by status (active, inactive, all)
- `created_after` (query) - Filter users created after this date (ISO 8601)
- `created_before` (query) - Filter users created before this date (ISO 8601)
- `X-Request-ID` (header) - Unique request identifier for tracking

**Responses:**

- `422` - Validation error in query parameters
- `500` - Internal server error
- `200` - Search results with pagination
- `400` - Invalid query parameters
- `401` - Unauthorized - invalid or missing token

### /health

#### GET

**Summary:** Health check

**Description:** Returns the health status of the API

**Responses:**

- `503` - API is unhealthy
- `200` - API is healthy

### /auth/logout

#### POST

**Summary:** User logout

**Description:** Invalidate the current access token

**Responses:**

- `200` - Logout successful
- `401` - Unauthorized


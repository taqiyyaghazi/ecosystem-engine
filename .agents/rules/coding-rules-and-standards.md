---
trigger: always_on
---

# Coding Rules & Standards

## HTTP Handlers
- **Request Binding**: Use `httputil.BindJSON(c, &req)` for JSON bodies and `httputil.BindQuery(c, &req)` for query parameters. These utilities provide standardized validation error formatting and type checking.
    - Avoid using `c.ShouldBindJSON` or `c.ShouldBindQuery` directly in handlers.
- **Error Handling**: Use `httputil.HandleError(c, err)` to return error responses. This ensures consistent error formats (e.g., standardized 400/401/404/500 responses) across all features.
    - Avoid using `c.JSON(http.Status..., gin.H{"error": ...})` manually.
- **Success Responses**: Always use `httputil.NewSuccessResponse(c, status, message, data)` for successful responses.
    - Avoid using `c.JSON(http.StatusOK, ...)` manually.
- **User Identification**: Always use `httputil.ExtractUserID(c)` to retrieve the authenticated user's ID from the session context. This utility automatically handles 401 Unauthorized responses if the session is missing or invalid.
    - Avoid manual extraction from `c.Get("session")` or `c.Get("user_id")` in handlers.

## Request Data Transfer Objects (DTOs)
- **Detailed Validation**: Do not just use `binding:"required"`. Use specific validation tags to enforce business logic at the DTO level:
    - Numbers: `gt=0`, `min=1`, `gte=0`.
    - Strings: `min=5`, `max=255`, `uuid`, `email`.
    - Coordinates: `latitude`, `longitude`.
- **Error Messages**: Ensure any new validation tags are also handled in `internal/apperror/validation.go` to provide human-readable error messages.

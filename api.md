# API Documentation

**Auth:** endpoints marked with `[JWT]` require `Authorization: Bearer <token>` header

## Account

---

### Register

`POST /account/register`

Creates a new user. Account 'active' status is 'false' until email is confirmed.

**Request body:**
```json
{
  "login": "string",
  "mail": "string",
  "password": "string",
  "description": "string"
}
```

**Responses:**

| Code | Body                               | Description                                      |
|------|------------------------------------|--------------------------------------------------|
| 201 | `{}`                               | Account created, confirmation code sent to email |
| 400 | `{"error": "validation error"}`    | Validation error                                 |
| 409 | `{"error": "login already taken"}` | login already taken                              |
| 409 | `{"error": "email already taken"}` | email already taken                              |
| 500 | `{"error": "server error"}`        |  Server error                                    |

---

### Send confirmation code

`POST /account/sendCode`

Sends a 6-digit confirmation code to the specified email.

**Request body:**
```json
{
  "mail": "string"
}
```

**Responses:**

| Code | Body                            | Description      |
|------|---------------------------------|------------------|
| 200 | `{"code": "string"}`            | Code sent        |
| 400 | `{"error": "validation error"}` | Validation error |
| 500 | `{"error": "server error"}`      |  Server error    |

---

### Confirm code

`POST /account/confirmCode`

Verifies the code and activates the account. Returns JWT on success.

**Request body:**
```json
{
  "email": "string",
  "code": "string"
}
```

**Responses:**

| Code | Body                                              | Description |
|------|---------------------------------------------------|-------------|
| 200 | `{ "token": "string", "refresh_token: "string" }` | Account activated |
| 400 | `{ "error": "Validation error" }`                              | Validation error |
| 400 | `{ "error": "Invalid or expired code" }`                              | Invalid or expired code |
| 500 | `{ "error": "Server error" }`                              | Server error |

---

### Login

`POST /account/auth`

Authenticates user by login and password. Returns JWT.

**Request body:**
```json
{
  "login": "string",
  "password": "string"
}
```

**Responses:**

| Code | Body                                              | Description |
|------|---------------------------------------------------|-------------|
| 200 | `{ "token": "string", "refresh_token: "string" }` | Success |
| 400 | `{ "error": "validation error" }`                 | Validation error |
| 401 | `{ "error": "invalid login or password" }`        | Invalid login or password |
| 403 | `{ "error": "account is not confirmed" }`         | Account is not confirmed |
| 500 | `{ "error": "server error" }`                     | Server error |

---

### Get account

`GET /account/getAccount`

Returns public account data by username.

**Responses:**

| Code | Body                               | Description |
|------|------------------------------------|-------------|
| 200 | account object                     | Success |
| 404 | `{ "error": "account not found" }` | Account not found |
| 500 | `{ "error": "server error" }`      | Server error |
| 400 | `{ "error": "validation error" }` | Validation error |

**Response body (200):**
```json
{
  "id": 0,
  "login": "string",
  "avatar": "string",
  "official": "string",
  "description": "string",
  "rating": 0,
  "followers": 0,
  "followed": 0
}
```

---

### Update account `[JWT]`

`POST /account/changeData`

Updates profile data of the authenticated user.

**Request body** (all fields optional):
```json
{
  "name": "string",
  "description": "string",
  "avatar": "file"
}
```

**Responses:**

| Code | Body                              | Description      |
|------|-----------------------------------|------------------|
| 200 | `{}`                              | Account updated  |
| 400 | `{ "error": "validation error" }` | Validation error |
| 401 | `{ "error": "Unauthorized" }`     | Unauthorized     |
| 500 | `{ "error": "server error" }`     |  Server error    |

---

### Delete account `[JWT]`

`DELETE /account/delete`

Soft deletes the account. Marks it as `deleted`, does not remove from database.

**Responses:**

| Code | Body                              | Description      |
|------|-----------------------------------|------------------|
| 200 | `{}`                              | Account deleted  |
| 400 | `{ "error": "validation error" }` | Validation error |
| 401 | `{ "error": "Unauthorized" }`     | Unauthorized     |
| 500 | `{ "error": "server error" }`     |  Server error    |

---

### Follow `[JWT]`

`POST /account/follow/:id`

Follow a user. Increments follower count for target, following count for current user.

**Responses:**

| Code | Body                                    | Description            |
|------|-----------------------------------------|------------------------|
| 200 | `{}`                                    | Followed               |
| 400 | `{ "error": "already following" }`      | Already following      |
| 400 | `{ "error": "cannot follow yourself" }` | cannot follow yourself |
| 401 | `{ "error": "Unauthorized" }`           | Unauthorized           |
| 404 | `{ "error": "account not found" }`      | Account not found      |
| 500 | `{ "error": "server error" }`           |  Server error          |

---

### Unfollow `[JWT]`

`POST /account/unfollow/:id`

Unfollow a user. Decrements follower count for target, following count for current user.

**Responses:**

| Code | Body | Description       |
|------|-------|-------------------|
| 200 | `{}`                               | Unfollowed        |
| 400 | `{ "error": "not following" }`     | not following     |
| 401 | `{ "error": "Unauthorized" }`      | Unauthorized      |
| 404 | `{ "error": "account not found" }` | Account not found |
| 500 | `{ "error": "server error" }`      | Server error      |

---

### Change password

`POST /account/changepassword`

Sends a confirmation code to the current email to initiate password change or reset.

**Request body:**
```json
{
  "mail": "string"
}
```

**Responses:**

| Code | body                              | Description        |
|------|-----------------------------------|--------------------|
| 200 | `{}`                              | Code sent to email |
| 400 | `{ "error": "validation error" }` | Validation error   |
| 500 | `{ "error": "server error" }`     |  Server error      |

---

### Confirm password change

`POST /account/confirmchangepassword`

Verifies the code and updates the password.

**Request body:**
```json
{
  "code": "string",
  "new_password": "string",
  "mail": "string"
}
```

**Responses:**

| Code | Body                                   | Description             |
|------|----------------------------------------|-------------------------|
| 200 | `{}`                                   | Password updated        |
| 400 | `{"error": "validation error"}`        | Validation error        |
| 400 | `{"error": "invalid or expired code"}` | Invalid or expired code |
| 500 | `{"error": "server error"}`            |  Server error           |

---

### Change email `[JWT]`

`POST /account/changeemail`

Sends a confirmation code to the current email to initiate email change.

**Responses:**

| Code | Body                            | Description                |
|------|---------------------------------|----------------------------|
| 200 | `{}`                            | Code sent to current email |
| 400 | `{"error": "validation error"}` | Validation error           |
| 401 | `{"error": "Unauthorized"}`     | Unauthorized               |
| 500 | `{"error": "server error"}`     |  Server error              |

---

### Confirm email change `[JWT]`

`POST /account/confirmchangeemail`

Verifies the code from current email and updates to new email. Requires password confirmation.

**Request body:**
```json
{
  "new_ mail": "string",
  "code": "string",
  "password": "string"
}
```

**Responses:**

| Code | Body                                  | Description             |
|------|---------------------------------------|-------------------------|
| 200 | `{}`                                  | Email updated           |
| 400 | `{"error": "validation error"}`       | Validation error        |
| 400 | `{"error": "invalid or expired code}` | Invalid or expired code |
| 401 | `{"error": "invalid password}`        | Invalid password        |
| 401 | `{"error": "Unauthorized}`            | Unauthorized            |
| 500 | `{"error": "server error}`            |  Server error           |

---

### Request verification `[JWT]`

`POST /account/requestverification`

Submits a request for a verified badge. Sent for admin review.

**Responses:**

| Code | Body                                     | Description               |
|------|------------------------------------------|---------------------------|
| 201  | `{}`                                     | Request submitted         |
| 400  | `{"error": "request already submitted}`  | Request already submitted |
| 400  | `{"error": "validation error"}`          | Validation error        |
| 401  | `{"error": "Unauthorized}`               | Unauthorized              |
| 500  | `{"error": "server error}`               |  Server error             |

### delete all sessions `[JWT]`

`DELETE /account/deletesessions`

delete sessions from all devices

**Responses**

| Code | Body                            | Description                |
|------ |---------------------------------|----------------------------|
| 200 | `{}`                            | alss sessions closed       |
| 400 | `{"error": "validation error"}` | Validation error           |
| 401 | `{"error": "Unauthorized}`      | Unauthorized              |
| 500 | `{"error": "server error}`      |  Server error             |
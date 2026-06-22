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

| Code | Description |
|------|-------------|
| 201 | Account created, confirmation code sent to email |
| 400 | Validation error |
| 409 | Email already taken |
| 409 | Username already taken |
| 500 | Server error |

---

### Send confirmation code

`POST /account/sendcode`

Sends a 6-digit confirmation code to the specified email.

**Request body:**
```json
{
  "email": "string"
}
```

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Code sent |
| 400 | Validation error |
| 500 | Server error |

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

| Code | Body | Description |
|------|------|-------------|
| 200 | `{ "token": "string" }` | Account activated |
| 400 | `{ "error": "..." }` | Validation error |
| 400 | `{ "error": "..." }` | Invalid or expired code |
| 500 | `{ "error": "..." }` | Server error |

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

| Code | Body | Description |
|------|------|-------------|
| 200 | `{ "token": "string" }` | Success |
| 400 | `{ "error": "..." }` | Validation error |
| 401 | `{ "error": "..." }` | Invalid login or password |
| 403 | `{ "error": "..." }` | Account is not confirmed |
| 500 | `{ "error": "..." }` | Server error |

---

### Get account

`GET /account/getAccount`

Returns public account data by username.

**Responses:**

| Code | Body | Description |
|------|------|-------------|
| 200 | account object | Success |
| 404 | `{ "error": "..." }` | Account not found |
| 500 | `{ "error": "..." }` | Server error |

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
  "bio": "string",
  "avatar": "string"
}
```

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Account updated |
| 400 | Validation error |
| 401 | Unauthorized |
| 500 | Server error |

---

### Delete account `[JWT]`

`DELETE /account/delete`

Soft deletes the account. Marks it as `deleted`, does not remove from database.

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Account deleted |
| 400 | Validation error |
| 401 | Unauthorized |
| 500 | Server error |

---

### Follow `[JWT]`

`POST /account/follow/:id`

Follow a user. Increments follower count for target, following count for current user.

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Followed |
| 400 | Already following / cannot follow yourself |
| 401 | Unauthorized |
| 404 | Account not found |
| 500 | Server error |

---

### Unfollow `[JWT]`

`OST /account/unfollow/:id`

Unfollow a user. Decrements follower count for target, following count for current user.

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Unfollowed |
| 400 | Not following |
| 401 | Unauthorized |
| 404 | Account not found |
| 500 | Server error |

---

### Change password `[JWT]`

`POST /account/changepassword`

Sends a confirmation code to the current email to initiate password change.

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Code sent to email |
| 401 | Unauthorized |
| 500 | Server error |

---

### Confirm password change `[JWT]`

`POST /account/confirmchangepassword`

Verifies the code and updates the password.

**Request body:**
```json
{
  "code": "string",
  "new_password": "string"
}
```

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Password updated |
| 400 | Validation error |
| 400 | Invalid or expired code |
| 401 | Unauthorized |
| 500 | Server error |

---

### Change email `[JWT]`

`POST /account/changeemail`

Sends a confirmation code to the current email to initiate email change.

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Code sent to current email |
| 400 | Validation error |
| 401 | Unauthorized |
| 500 | Server error |

---

### Confirm email change `[JWT]`

`POST /account/confirmchangeemail`

Verifies the code from current email and updates to new email. Requires password confirmation.

**Request body:**
```json
{
  "mail": "string",
  "code": "string",
  "password": "string"
}
```

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Email updated |
| 400 | Validation error |
| 400 | Invalid or expired code |
| 401 | Invalid password |
| 401 | Unauthorized |
| 500 | Server error |

---

### Request verification `[JWT]`

`POST /account/requestverification`

Submits a request for a verified badge. Sent for admin review.

**Responses:**

| Code | Description |
|------|-------------|
| 200 | Request submitted |
| 400 | Request already submitted |
| 401 | Unauthorized |
| 500 | Server error |
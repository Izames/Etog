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

## Event

---

### Create event `[JWT]`

`POST /event/create`

Creates a new event. The authenticated user becomes the owner of the event.

**Request body:**

```json
{
  "name": "string",
  "price": 0.00,
  "address": "string",
  "city": "string",
  "latitude": 1.000000,
  "longitude": 1.0000000,
  "type": "string",
  "date": "2026-08-13T18:30:00+03:00",
  "description": "string",
  "passed": "string",
  "max_people": 10,
  "media": "files",
  "phone": "string",
  "mail": "string",
  "telegram": "string"
}
```
**Responses:**

| Code | Body                              | Description      |
| ---- | --------------------------------- | ---------------- |
| 201  |                    | Event created    |
| 400  | `{ "error": "validation error" }` | Validation error |
| 401  | `{ "error": "Unauthorized" }`     | Unauthorized     |
| 500  | `{ "error": "server error" }`     | Server error     |

---

### Get event

`GET /event/:id`

Returns event data by event ID.

**Responses:**

| Code | Body                             | Description     |
| ---- | -------------------------------- | --------------- |
| 200  | `event object`                   | Success         |
| 404  | `{ "error": "event not found" }` | Event not found |
| 500  | `{ "error": "server error" }`    | Server error    |

**Response body (200):**

```json
{
  "id": 0,
  "organizer": 0,
  "organizer_name": "string",
  "name": "string",
  "description": "string",
  "type": "string",
  "date": "2026-08-20T18:00:00Z",
  "latitude": 1.000000,
  "longitude": 1.0000000,
  "city": "string",
  "address": "string",
  "price": 1.0,
  "passed": "string",
  "max_people": 100,
  "people": 0,
  "media": ["string", "string"],
  "phone": "string",
  "mail": "string",
  "telegram": "string"
}
```
 ---

### Get events

`GET /event`

Returns a paginated list of events.

**Query parameters:**

```text
?page=1
&limit=20
&user_id=123
&type=concert
&city=Chelyabinsk
&sort=date
&order=asc
&subscribed_only=true
```

**Available sort parameters:**

* `date` — sort by event date

**Available order parameters:**

* `asc` — ascending
* `desc` — descending


**Responses:**

| Code | Body                              | Description                   |
| ---- | --------------------------------- | ----------------------------- |
| 200  | `event list`                      | Events returned               |
| 400  | `{ "error": "validation error" }` | Invalid pagination parameters |
| 500  | `{ "error": "server error" }`     | Server error                  |

**Response body (200):**

```json
{
  "events": [
    {
      "id": 0,
      "organizer": 0,
      "organizer_name": "string",
      "name": "string",
      "description": "string",
      "type": "string",
      "date": "2026-08-20T18:00:00Z",
      "latitude": 1.000000,
      "longitude": 1.0000000,
      "city": "string",
      "address": "string",
      "price": 1.0,
      "passed": "string",
      "max_people": 100,
      "people": 0,
      "media": ["string", "string"],
      "phone": "string",
      "mail": "string",
      "telegram": "string"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```

---

### Join event

`POST /event/join/:id`

Subscribes the authenticated user to an event.

**Responses:**

| Code | Body                                | Description                |
| ---- | ----------------------------------- | -------------------------- |
| 200  | `{}`                                | Subscribed                 |
| 400  | `{ "error": "already subscribed" }` | User is already subscribed |
| 401  | `{ "error": "Unauthorized" }`       | Unauthorized               |
| 404  | `{ "error": "event not found" }`    | Event not found            |
| 500  | `{ "error": "server error" }`       | Server error               |

---

### Unjoin event

`POST /event/unjoin/:id`

Unsubscribes the authenticated user from an event.

**Responses:**

| Code | Body                             | Description            |
| ---- | -------------------------------- | ---------------------- |
| 200  | `{}`                             | Unsubscribed           |
| 400  | `{ "error": "not subscribed" }`  | User is not subscribed |
| 401  | `{ "error": "Unauthorized" }`    | Unauthorized           |
| 404  | `{ "error": "event not found" }` | Event not found        |
| 500  | `{ "error": "server error" }`    | Server error           |

---

### Update event

`POST /event/changeData/:id`

Updates event data. Only the event creator can update the event.

**Request body** (all fields optional):

```json
{
  "name": "string",
  "price": 0.00,
  "address": "string",
  "city": "string",
  "latitude": 1.000000,
  "longitude": 1.0000000,
  "type": "string",
  "date": "2026-08-13T18:30:00+03:00",
  "description": "string",
  "passed": "string",
  "max_people": 10,
  "media": "files",
  "phone": "string",
  "mail": "string",
  "telegram": "string"
}
```

**Responses:**

| Code | Body                              | Description                 |
| ---- | --------------------------------- | --------------------------- |
| 200  | `{}`                              | Event updated               |
| 400  | `{ "error": "validation error" }` | Validation error            |
| 401  | `{ "error": "Unauthorized" }`     | Unauthorized                |
| 403  | `{ "error": "not event owner" }`  | User is not the event owner |
| 404  | `{ "error": "event not found" }`  | Event not found             |
| 500  | `{ "error": "server error" }`     | Server error                |

---

### Delete event

`DELETE /event/:id`

Soft deletes an event. Only the event creator can delete the event.

**Responses:**

| Code | Body                             | Description                 |
| ---- | -------------------------------- | --------------------------- |
| 200  | `{}`                             | Event deleted               |
| 401  | `{ "error": "Unauthorized" }`    | Unauthorized                |
| 403  | `{ "error": "not event owner" }` | User is not the event owner |
| 404  | `{ "error": "event not found" }` | Event not found             |
| 500  | `{ "error": "server error" }`    | Server error                |

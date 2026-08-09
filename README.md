 <h1>Migrator:</h1>
<h2>Migrator flags:</h2>
<ul>
  <li>--conn-str="" [connect-string to database]</li>
  <li>--path=""     [path to migration files]</li>
  <li>--table=""    [migration table name]</li>
  <li>--force=0     [forceful migration to clean]</li>
  <li>--down=true   [down migration]</li>
  <li>--version=0   [version choose]</li>
</ul>

<h2>Migrator сommands examples:</h2>
<ul>
  <li>do migrations:                     go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations"</li>
  <li>do migrations and name migr table: go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --table="migr-table"</li>
  <li>force dirty version:               go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --force=1</li>
  <li>down to version:                   go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --version=1</li>
  <li>down for x steps:                  go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --down=true --steps=1</li>
</ul>

<h1>Functions</h1>
<h2>Account</h2>

<h3>Register</h3>
<p>Create new user. Account is inactive until email is confirmed.</p>
<p>JWT: not required</p>
<ul>
  <li>201 — account created, code sent to email</li>
  <li>400 — validation error</li>
  <li>409 — email already taken</li>
  <li>409 — username already taken</li>
  <li>500 — server error</li>
</ul>

<h3>Send confirmation code</h3>
<p>Send 6-digit confirmation code to the specified email.</p>
<p>JWT: not required</p>
<ul>
  <li>200 — code sent</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Confirm code</h3>
<p>Verify the code and activate the account. Returns JWT on success.</p>
<p>JWT: not required</p>
<ul>
  <li>200 — account activated, returns JWT</li>
  <li>400 — validation error</li>
  <li>400 — invalid or expired code</li>
  <li>500 — server error</li>
</ul>

<h3>Login</h3>
<p>Authenticate user by login/email and password. Returns JWT.</p>
<p>JWT: not required</p>
<ul>
  <li>200 — returns JWT</li>
  <li>400 — validation error</li>
  <li>401 — invalid login or password</li>
  <li>403 — account is not confirmed</li>
  <li>500 — server error</li>
</ul>

<h3>Get account</h3>
<p>Get public account data by username or id.</p>
<p>JWT: not required</p>
<ul>
  <li>200 — returns account data</li>
  <li>404 — account not found</li>
  <li>500 — server error</li>
</ul>

<h3>Update account</h3>
<p>Update profile data: name, avatar, bio, etc.</p>
<p>JWT: required</p>
<ul>
  <li>200 — account updated</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Delete account</h3>
<p>Soft delete account. Account is marked as deleted, not removed from database.</p>
<p>JWT: required</p>
<ul>
  <li>200 — account deleted</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Follow</h3>
<p>Follow a user. Increments follower count for target and following count for current user.</p>
<p>JWT: required</p>
<ul>
  <li>200 — followed</li>
  <li>400 — validation error (e.g. already following, following yourself)</li>
  <li>500 — server error</li>
</ul>

<h3>Unfollow</h3>
<p>Unfollow a user. Decrements follower count for target and following count for current user.</p>
<p>JWT: required</p>
<ul>
  <li>200 — unfollowed</li>
  <li>400 — validation error (e.g. not following)</li>
  <li>500 — server error</li>
</ul>

<h3>Change password</h3>
<p>Request password change. Sends confirmation code to current email.</p>
<p>JWT: required</p>
<ul>
  <li>200 — code sent to email</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Confirm password change</h3>
<p>Verify code and update password.</p>
<p>JWT: required</p>
<ul>
  <li>200 — password updated</li>
  <li>400 — validation error</li>
  <li>400 — invalid or expired code</li>
  <li>500 — server error</li>
</ul>

<h3>Change email</h3>
<p>Request email change. Sends confirmation code to current email.</p>
<p>JWT: required</p>
<ul>
  <li>200 — code sent to current email</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Confirm email change</h3>
<p>Verify code from current email and update to new email. Requires password confirmation.</p>
<p>JWT: required</p>
<ul>
  <li>200 — email updated</li>
  <li>400 — validation error</li>
  <li>400 — invalid or expired code</li>
  <li>401 — invalid password</li>
  <li>500 — server error</li>
</ul>

<h3>Request verification</h3>
<p>Submit a request for a verified badge. Sent for admin review.</p>
<p>JWT: required</p>
<ul>
  <li>200 — request submitted</li>
  <li>400 — validation error (e.g. request already submitted)</li>
  <li>500 — server error</li>
</ul>

<h2>Event</h2>

<h3>Create event</h3>
<p>Create a new event.</p>
<p>JWT: required</p>
<ul>
  <li>201 — event created</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Get event</h3>
<p>Get event data by id.</p>
<p>JWT: not required</p>
<ul>
  <li>200 — returns event data</li>
  <li>404 — event not found</li>
  <li>500 — server error</li>
</ul>

<h3>Get events</h3>
<p>Get list of events with pagination.</p>
<p>JWT: not required</p>
<ul>
  <li>200 — returns list of events</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Search event by user</h3>
<p>Get list of events created by a specific user.</p>
<p>JWT: not required</p>
<ul>
  <li>200 — returns list of events</li>
  <li>400 — validation error</li>
  <li>404 — user not found</li>
  <li>500 — server error</li>
</ul>

<h3>Filter events</h3>
<p>Get list of events filtered by parameters (category, date range, location, etc).</p>
<p>JWT: not required</p>
<ul>
  <li>200 — returns filtered list of events</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Sort events</h3>
<p>Get list of events sorted by parameter (date, popularity, etc).</p>
<p>JWT: not required</p>
<ul>
  <li>200 — returns sorted list of events</li>
  <li>400 — validation error</li>
  <li>500 — server error</li>
</ul>

<h3>Subscribed events</h3>
<p>Get list of events the current user is subscribed to.</p>
<p>JWT: required</p>
<ul>
  <li>200 — returns list of events</li>
  <li>500 — server error</li>
</ul>

<h3>join to event</h3>
<p>Subscribe current user to an event.</p>
<p>JWT: required</p>
<ul>
  <li>200 — subscribed</li>
  <li>400 — validation error (e.g. already subscribed)</li>
  <li>404 — event not found</li>
  <li>500 — server error</li>
</ul>

<h3>Unjoin from event</h3>
<p>Unsubscribe current user from an event.</p>
<p>JWT: required</p>
<ul>
  <li>200 — unsubscribed</li>
  <li>400 — validation error (e.g. not subscribed)</li>
  <li>404 — event not found</li>
  <li>500 — server error</li>
</ul>

<h3>Update event</h3>
<p>Update event data. Only the creator can update the event.</p>
<p>JWT: required</p>
<ul>
  <li>200 — event updated</li>
  <li>400 — validation error</li>
  <li>403 — not the event owner</li>
  <li>404 — event not found</li>
  <li>500 — server error</li>
</ul>

<h3>Delete event</h3>
<p>Soft delete event. Only the creator can delete the event.</p>
<p>JWT: required</p>
<ul>
  <li>200 — event deleted</li>
  <li>403 — not the event owner</li>
  <li>404 — event not found</li>
  <li>500 — server error</li>
</ul>
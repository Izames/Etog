DROP TABLE IF EXISTS account;
DROP INDEX IF EXISTS account_login;


DROP TABLE IF EXISTS subscribes;
DROP INDEX IF EXISTS subscribers_on_account;
DROP INDEX IF EXISTS accounts_subscribers;
DROP INDEX IF EXISTS idx_current_subscribe;
DROP INDEX IF EXISTS idx_subscribe_unique;

DROP TABLE IF EXISTS event;

DROP INDEX IF EXISTS idx_event_organizer;
DROP INDEX IF EXISTS idx_event_name;
DROP INDEX IF EXISTS idx_event_price;
DROP INDEX IF EXISTS idx_event_city;
DROP INDEX IF EXISTS idx_event_vector;
DROP INDEX IF EXISTS idx_event_type;
DROP INDEX IF EXISTS idx_event_date;
DROP INDEX IF EXISTS idx_event_max_people;
DROP INDEX IF EXISTS idx_event_passed;
--прикинуть индекс cords когда сайт станет доступен



DROP TABLE IF EXISTS contact;

DROP TABLE IF EXISTS joined_on_event;

DROP INDEX IF EXISTS idx_joined_event;
DROP INDEX IF EXISTS idx_on_event_join;
DROP INDEX IF EXISTS idx_current_join;
DROP INDEX IF EXISTS idx_join_unique;

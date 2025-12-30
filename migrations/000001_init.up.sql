CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name TEXT NOT NULL,
    cost_rub INTEGER NOT NULL CHECK (cost_rub > 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE
);
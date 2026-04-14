CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wishlists (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title        TEXT        NOT NULL,
    description  TEXT,
    event_date   DATE        NOT NULL,
    public_token UUID        UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_wishlists_user_id     ON wishlists (user_id);
CREATE INDEX idx_wishlists_public_token ON wishlists (public_token) WHERE public_token IS NOT NULL;

CREATE TABLE items (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    wishlist_id UUID        NOT NULL REFERENCES wishlists(id) ON DELETE CASCADE,
    title       TEXT        NOT NULL,
    description TEXT,
    product_url TEXT,
    priority    INTEGER     NOT NULL,
    reserved    BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_items_wishlist_id ON items (wishlist_id);

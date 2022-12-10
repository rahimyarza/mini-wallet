create table wallet (
    xid varchar(255),
    token varchar(255),
    wid varchar(255),
    balance integer DEFAULT 0,
    is_enabled boolean DEFAULT FALSE
);
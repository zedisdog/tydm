CREATE TABLE bm_admin (
    id bigint NOT NULL IDENTITY,
    name varchar(255) NOT NULL,
    phone varchar(255) NOT NULL,
    password varchar(255) NOT NULL,
    role_id bigint DEFAULT 0 NOT NULL,
    status int NOT NULL,
    create_time int DEFAULT 0 NOT NULL,
    delete_time int DEFAULT 0 NOT NULL,
    CONSTRAINT bm_admin_PK PRIMARY KEY (id)
);
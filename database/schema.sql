BEGIN;

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE users_id_seq OWNER TO jump;

CREATE TABLE users (
    id INTEGER DEFAULT nextval('users_id_seq'::regclass) PRIMARY KEY,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	balance BIGINT NOT NULL
);

CREATE SEQUENCE invoices_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE invoices_id_seq OWNER TO jump;

CREATE TABLE invoice_status (
    status TEXT NOT NULL,
    CONSTRAINT invoice_status_unique UNIQUE (status)
);

INSERT INTO invoice_status VALUES
    ('pending'),
    ('paid');

CREATE TABLE invoices (
    id INTEGER DEFAULT nextval('invoices_id_seq'::regclass) PRIMARY KEY,
	user_id INTEGER REFERENCES users(id),
	status TEXT REFERENCES invoice_status (status) DEFAULT 'pending',
	label TEXT NOT NULL,
	amount BIGINT NOT NULL
);

INSERT INTO users (first_name, last_name, balance) VALUES
    ('Bob', 'Loco', 241817),
    ('Kevin', 'Findus', 49297),
    ('Lynne', 'Gwafranca', 82540),
    ('Art', 'Decco', 402758),
    ('Lynne', 'Gwistic', 226777),
    ('Polly', 'Ester Undawair', 144970),
    ('Oscar', 'Nommanee', 205387),
    ('Laura', 'Biding', 520060),
    ('Laura', 'Norda', 565074),
    ('Des', 'Ignayshun', 436180),
    ('Mike', 'Rowe-Soft', 818313),
    ('Anne', 'Kwayted', 189588),
    ('Wayde', 'Thabalanz',97005),
    ('Dee', 'Mandingboss', 276296),
    ('Sly', 'Meedentalfloss', 932505),
    ('Stanley', 'Knife', 500691),
    ('Wynn', 'Dozeaplikayshun', 478333);

COMMIT;

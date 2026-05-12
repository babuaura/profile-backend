create table if not exists profiles (
	id text primary key,
	data jsonb not null,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table if not exists contact_messages (
	id text primary key,
	name text not null,
	email text not null,
	budget text not null default '',
	message text not null,
	source text not null,
	status text not null,
	created_at timestamptz not null
);

create index if not exists contact_messages_created_at_idx
	on contact_messages (created_at desc);

create table if not exists personal_notes (
	id text primary key,
	title text not null,
	body text not null,
	tags jsonb not null default '[]'::jsonb,
	pinned boolean not null default false,
	created_at timestamptz not null,
	updated_at timestamptz not null
);

create index if not exists personal_notes_order_idx
	on personal_notes (pinned desc, updated_at desc);

create table if not exists personal_reminders (
	id text primary key,
	title text not null,
	notes text not null,
	due_at timestamptz not null,
	done boolean not null default false,
	created_at timestamptz not null,
	updated_at timestamptz not null
);

create index if not exists personal_reminders_order_idx
	on personal_reminders (done asc, due_at asc);

create table if not exists personal_transactions (
	id text primary key,
	type text not null check (type in ('income', 'expense')),
	amount numeric(12,2) not null check (amount > 0),
	category text not null,
	note text not null,
	occurred_at timestamptz not null,
	created_at timestamptz not null,
	updated_at timestamptz not null
);

create index if not exists personal_transactions_occurred_at_idx
	on personal_transactions (occurred_at desc);

create table if not exists personal_habits (
	id text primary key,
	name text not null,
	target text not null,
	frequency text not null default 'daily',
	streak integer not null default 0,
	last_checked_at timestamptz not null default '0001-01-01 00:00:00+00',
	created_at timestamptz not null,
	updated_at timestamptz not null
);

create index if not exists personal_habits_order_idx
	on personal_habits (streak desc, updated_at desc);

create table if not exists notification_tokens (
	id text primary key,
	token text not null unique,
	platform text not null,
	created_at timestamptz not null,
	updated_at timestamptz not null
);

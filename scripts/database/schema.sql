CREATE TABLE
  public.users (
    id character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    is_enabled boolean NOT NULL DEFAULT false,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    updated_at timestamp without time zone NOT NULL DEFAULT now()
  );

ALTER TABLE
  public.users
ADD
  CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX users_unique_email ON public."users" USING btree (email);

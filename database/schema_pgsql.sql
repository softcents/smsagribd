--
-- PostgreSQL database dump
--

\restrict R37nrIu0qgfQ5OMAnbJj5vogEayRzWKwMWr8BvpgOYwfueDxtrc3Jx5vMvs87u7

-- Dumped from database version 18.0
-- Dumped by pg_dump version 18.0

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: smsagribd_sc; Type: SCHEMA; Schema: -; Owner: smsagribd_sc
--

CREATE SCHEMA smsagribd_sc;


ALTER SCHEMA smsagribd_sc OWNER TO smsagribd_sc;

--
-- Name: bigint_bool_eq(bigint, boolean); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.bigint_bool_eq(i bigint, b boolean) RETURNS boolean
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT (i != 0) = b; $$;


ALTER FUNCTION smsagribd_sc.bigint_bool_eq(i bigint, b boolean) OWNER TO smsagribd_sc;

--
-- Name: bool_bigint_eq(boolean, bigint); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.bool_bigint_eq(b boolean, i bigint) RETURNS boolean
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT b = (i != 0); $$;


ALTER FUNCTION smsagribd_sc.bool_bigint_eq(b boolean, i bigint) OWNER TO smsagribd_sc;

--
-- Name: bool_int_eq(boolean, integer); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.bool_int_eq(b boolean, i integer) RETURNS boolean
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT b = (i != 0); $$;


ALTER FUNCTION smsagribd_sc.bool_int_eq(b boolean, i integer) OWNER TO smsagribd_sc;

--
-- Name: database(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.database() RETURNS character varying
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN 'smsagribd_sc';
END;
$$;


ALTER FUNCTION smsagribd_sc.database() OWNER TO smsagribd_sc;

--
-- Name: date_add(timestamp with time zone, interval); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.date_add(ts timestamp with time zone, iv interval) RETURNS timestamp with time zone
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN ts + iv;
END;
$$;


ALTER FUNCTION smsagribd_sc.date_add(ts timestamp with time zone, iv interval) OWNER TO smsagribd_sc;

--
-- Name: date_sub(timestamp with time zone, integer); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.date_sub(ts timestamp with time zone, minutes integer) RETURNS timestamp with time zone
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN ts - (minutes || ' minutes')::interval;
END;
$$;


ALTER FUNCTION smsagribd_sc.date_sub(ts timestamp with time zone, minutes integer) OWNER TO smsagribd_sc;

--
-- Name: date_sub(timestamp with time zone, interval); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.date_sub(ts timestamp with time zone, iv interval) RETURNS timestamp with time zone
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN ts - iv;
END;
$$;


ALTER FUNCTION smsagribd_sc.date_sub(ts timestamp with time zone, iv interval) OWNER TO smsagribd_sc;

--
-- Name: int_bool_eq(integer, boolean); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.int_bool_eq(i integer, b boolean) RETURNS boolean
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT (i != 0) = b; $$;


ALTER FUNCTION smsagribd_sc.int_bool_eq(i integer, b boolean) OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_access_tokens(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_access_tokens() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_access_tokens() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_batches(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_batches() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_batches() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_campaigns(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_campaigns() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_campaigns() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_contact_group(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_contact_group() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_contact_group() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_contacts(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_contacts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_contacts() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_countries(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_countries() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_countries() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_deposits(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_deposits() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_deposits() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_gateway_operator_rates(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_gateway_operator_rates() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_gateway_operator_rates() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_gateways(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_gateways() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_gateways() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_groups(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_groups() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_groups() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_ip_whitelist_audit_logs(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_ip_whitelist_audit_logs() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_ip_whitelist_audit_logs() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_ip_whitelist_requests(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_ip_whitelist_requests() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_ip_whitelist_requests() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_kyc_audit_logs(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_kyc_audit_logs() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_kyc_audit_logs() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_kyc_role_permissions(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_kyc_role_permissions() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_kyc_role_permissions() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_messages(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_messages() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_messages() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_operators(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_operators() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_operators() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_payment_webhooks(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_payment_webhooks() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_payment_webhooks() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_phone_photo_sessions(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_phone_photo_sessions() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_phone_photo_sessions() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_route_plan_gateways(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_route_plan_gateways() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_route_plan_gateways() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_route_plans(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_route_plans() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_route_plans() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_route_rules(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_route_rules() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_route_rules() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_sender_user(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_sender_user() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_sender_user() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_senders(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_senders() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_senders() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_settings(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_settings() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_settings() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_signup_slug_aliases(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_signup_slug_aliases() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_signup_slug_aliases() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_sms_logs(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_sms_logs() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_sms_logs() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_sp_distributions(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_sp_distributions() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_sp_distributions() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_tariff_plan_assignments(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_plan_assignments() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_plan_assignments() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_tariff_plans(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_plans() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_plans() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_tariff_rate_history(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_rate_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_rate_history() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_tariff_rates(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_rates() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_rates() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_templates(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_templates() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_templates() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_tickets(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_tickets() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_tickets() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_transactions(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_transactions() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_transactions() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_user_operator_rates(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_user_operator_rates() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_user_operator_rates() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_users(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_users() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_users() OWNER TO smsagribd_sc;

--
-- Name: on_update_current_timestamp_verification_otps(); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.on_update_current_timestamp_verification_otps() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


ALTER FUNCTION smsagribd_sc.on_update_current_timestamp_verification_otps() OWNER TO smsagribd_sc;

--
-- Name: sum_boolean_accum(bigint, boolean); Type: FUNCTION; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE FUNCTION smsagribd_sc.sum_boolean_accum(state bigint, b boolean) RETURNS bigint
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    IF b THEN
        RETURN state + 1;
    ELSE
        RETURN state;
    END IF;
END;
$$;


ALTER FUNCTION smsagribd_sc.sum_boolean_accum(state bigint, b boolean) OWNER TO smsagribd_sc;

--
-- Name: sum(boolean); Type: AGGREGATE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE AGGREGATE smsagribd_sc.sum(boolean) (
    SFUNC = smsagribd_sc.sum_boolean_accum,
    STYPE = bigint,
    INITCOND = '0'
);


ALTER AGGREGATE smsagribd_sc.sum(boolean) OWNER TO smsagribd_sc;

--
-- Name: =; Type: OPERATOR; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE OPERATOR smsagribd_sc.= (
    FUNCTION = smsagribd_sc.bool_int_eq,
    LEFTARG = boolean,
    RIGHTARG = integer,
    COMMUTATOR = OPERATOR(smsagribd_sc.=)
);


ALTER OPERATOR smsagribd_sc.= (boolean, integer) OWNER TO smsagribd_sc;

--
-- Name: =; Type: OPERATOR; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE OPERATOR smsagribd_sc.= (
    FUNCTION = smsagribd_sc.bool_bigint_eq,
    LEFTARG = boolean,
    RIGHTARG = bigint,
    COMMUTATOR = OPERATOR(smsagribd_sc.=)
);


ALTER OPERATOR smsagribd_sc.= (boolean, bigint) OWNER TO smsagribd_sc;

--
-- Name: =; Type: OPERATOR; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE OPERATOR smsagribd_sc.= (
    FUNCTION = smsagribd_sc.int_bool_eq,
    LEFTARG = integer,
    RIGHTARG = boolean,
    COMMUTATOR = OPERATOR(smsagribd_sc.=)
);


ALTER OPERATOR smsagribd_sc.= (integer, boolean) OWNER TO smsagribd_sc;

--
-- Name: =; Type: OPERATOR; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE OPERATOR smsagribd_sc.= (
    FUNCTION = smsagribd_sc.bigint_bool_eq,
    LEFTARG = bigint,
    RIGHTARG = boolean,
    COMMUTATOR = OPERATOR(smsagribd_sc.=)
);


ALTER OPERATOR smsagribd_sc.= (bigint, boolean) OWNER TO smsagribd_sc;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: access_tokens; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.access_tokens (
    id bigint NOT NULL,
    tokenable_type character varying(191) NOT NULL,
    tokenable_id numeric NOT NULL,
    name character varying(191) NOT NULL,
    token character varying(64) NOT NULL,
    abilities text,
    last_used_at timestamp with time zone,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.access_tokens OWNER TO smsagribd_sc;

--
-- Name: access_tokens_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.access_tokens_id_seq OWNER TO smsagribd_sc;

--
-- Name: access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.access_tokens_id_seq OWNED BY smsagribd_sc.access_tokens.id;


--
-- Name: batches; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.batches (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    sender_id numeric,
    body text NOT NULL,
    source_type character varying(64) DEFAULT 'direct'::character varying NOT NULL,
    source_ip character varying(64) DEFAULT NULL::character varying,
    source_config text,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    approval character varying(32) DEFAULT 'approved'::character varying NOT NULL,
    total_recipients bigint DEFAULT '0'::bigint NOT NULL,
    valid_count bigint DEFAULT '0'::bigint NOT NULL,
    invalid_count bigint DEFAULT '0'::bigint NOT NULL,
    failed_report_path character varying(255) DEFAULT NULL::character varying,
    scheduled_at timestamp with time zone,
    completed_at timestamp with time zone,
    last_processed_cursor numeric DEFAULT '0'::numeric NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.batches OWNER TO smsagribd_sc;

--
-- Name: batches_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.batches_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.batches_id_seq OWNER TO smsagribd_sc;

--
-- Name: batches_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.batches_id_seq OWNED BY smsagribd_sc.batches.id;


--
-- Name: cache; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.cache (
    key character varying(255) NOT NULL,
    value text NOT NULL,
    expiration bigint NOT NULL
);


ALTER TABLE smsagribd_sc.cache OWNER TO smsagribd_sc;

--
-- Name: campaigns; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.campaigns (
    id bigint NOT NULL,
    user_id numeric DEFAULT '0'::numeric NOT NULL,
    title character varying(255) DEFAULT NULL::character varying,
    status character varying(50) DEFAULT 'draft'::character varying NOT NULL,
    scheduled_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.campaigns OWNER TO smsagribd_sc;

--
-- Name: campaigns_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.campaigns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.campaigns_id_seq OWNER TO smsagribd_sc;

--
-- Name: campaigns_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.campaigns_id_seq OWNED BY smsagribd_sc.campaigns.id;


--
-- Name: contact_group; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.contact_group (
    id bigint NOT NULL,
    contact_id numeric NOT NULL,
    group_id numeric NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.contact_group OWNER TO smsagribd_sc;

--
-- Name: contact_group_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.contact_group_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.contact_group_id_seq OWNER TO smsagribd_sc;

--
-- Name: contact_group_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.contact_group_id_seq OWNED BY smsagribd_sc.contact_group.id;


--
-- Name: contacts; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.contacts (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    name character varying(255) DEFAULT ''::character varying NOT NULL,
    phone character varying(32) NOT NULL,
    email character varying(255) DEFAULT NULL::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.contacts OWNER TO smsagribd_sc;

--
-- Name: contacts_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.contacts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.contacts_id_seq OWNER TO smsagribd_sc;

--
-- Name: contacts_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.contacts_id_seq OWNED BY smsagribd_sc.contacts.id;


--
-- Name: countries; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.countries (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    code character varying(10) NOT NULL,
    iso_alpha2 character varying(10) NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.countries OWNER TO smsagribd_sc;

--
-- Name: countries_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.countries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.countries_id_seq OWNER TO smsagribd_sc;

--
-- Name: countries_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.countries_id_seq OWNED BY smsagribd_sc.countries.id;


--
-- Name: deposits; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.deposits (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    deposited_by_id numeric,
    amount numeric(15,4) DEFAULT 0.0000 NOT NULL,
    currency character varying(10) DEFAULT 'BDT'::character varying NOT NULL,
    payment_method character varying(64) DEFAULT 'manual'::character varying NOT NULL,
    transaction_id character varying(100) DEFAULT NULL::character varying,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    paid_at timestamp with time zone,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.deposits OWNER TO smsagribd_sc;

--
-- Name: deposits_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.deposits_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.deposits_id_seq OWNER TO smsagribd_sc;

--
-- Name: deposits_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.deposits_id_seq OWNED BY smsagribd_sc.deposits.id;


--
-- Name: gateway_operator_rates; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.gateway_operator_rates (
    id bigint NOT NULL,
    gateway_id numeric NOT NULL,
    operator_id numeric NOT NULL,
    masking_price numeric(10,4) DEFAULT 0.0000 NOT NULL,
    non_masking_price numeric(10,4) DEFAULT 0.0000 NOT NULL,
    dipping_charge numeric(10,4) DEFAULT 0.0000 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.gateway_operator_rates OWNER TO smsagribd_sc;

--
-- Name: gateway_operator_rates_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.gateway_operator_rates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.gateway_operator_rates_id_seq OWNER TO smsagribd_sc;

--
-- Name: gateway_operator_rates_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.gateway_operator_rates_id_seq OWNED BY smsagribd_sc.gateway_operator_rates.id;


--
-- Name: gateways; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.gateways (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    driver character varying(64) DEFAULT 'log'::character varying NOT NULL,
    credentials text NOT NULL,
    requests_per_second bigint DEFAULT '10'::bigint NOT NULL,
    cost_mode character varying(32) DEFAULT 'dynamic'::character varying NOT NULL,
    is_enabled boolean DEFAULT true NOT NULL,
    is_otp_enabled boolean DEFAULT false NOT NULL,
    priority bigint DEFAULT '0'::bigint NOT NULL,
    timeout_seconds bigint DEFAULT '30'::bigint NOT NULL,
    dipping_required boolean DEFAULT false NOT NULL,
    has_third_party_dipping boolean DEFAULT false NOT NULL,
    dlr_support boolean DEFAULT true NOT NULL,
    supports_bulk boolean DEFAULT true NOT NULL,
    health_status character varying(32) DEFAULT 'healthy'::character varying NOT NULL,
    last_health_check_at timestamp with time zone,
    last_health_latency_ms bigint DEFAULT '0'::bigint,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.gateways OWNER TO smsagribd_sc;

--
-- Name: gateways_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.gateways_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.gateways_id_seq OWNER TO smsagribd_sc;

--
-- Name: gateways_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.gateways_id_seq OWNED BY smsagribd_sc.gateways.id;


--
-- Name: groups; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.groups (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    name character varying(191) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.groups OWNER TO smsagribd_sc;

--
-- Name: groups_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.groups_id_seq OWNER TO smsagribd_sc;

--
-- Name: groups_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.groups_id_seq OWNED BY smsagribd_sc.groups.id;


--
-- Name: ip_whitelist_audit_logs; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.ip_whitelist_audit_logs (
    id bigint NOT NULL,
    ip_whitelist_request_id numeric NOT NULL,
    user_id numeric NOT NULL,
    actor_id numeric,
    actor_name character varying(191) DEFAULT NULL::character varying,
    actor_type character varying(64) DEFAULT 'admin'::character varying NOT NULL,
    action character varying(64) NOT NULL,
    ip_address character varying(45) NOT NULL,
    old_values text,
    new_values text,
    note text,
    source_ip character varying(45) DEFAULT NULL::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.ip_whitelist_audit_logs OWNER TO smsagribd_sc;

--
-- Name: ip_whitelist_audit_logs_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.ip_whitelist_audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.ip_whitelist_audit_logs_id_seq OWNER TO smsagribd_sc;

--
-- Name: ip_whitelist_audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.ip_whitelist_audit_logs_id_seq OWNED BY smsagribd_sc.ip_whitelist_audit_logs.id;


--
-- Name: ip_whitelist_requests; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.ip_whitelist_requests (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    ip_address character varying(45) NOT NULL,
    label character varying(191) DEFAULT NULL::character varying,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    admin_note text,
    reviewed_by numeric,
    reviewed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.ip_whitelist_requests OWNER TO smsagribd_sc;

--
-- Name: ip_whitelist_requests_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.ip_whitelist_requests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.ip_whitelist_requests_id_seq OWNER TO smsagribd_sc;

--
-- Name: ip_whitelist_requests_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.ip_whitelist_requests_id_seq OWNED BY smsagribd_sc.ip_whitelist_requests.id;


--
-- Name: kyc_audit_logs; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.kyc_audit_logs (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    admin_id numeric,
    admin_name character varying(191) DEFAULT NULL::character varying,
    action character varying(50) NOT NULL,
    field_name character varying(50) DEFAULT 'kyc_status'::character varying,
    old_value character varying(191) DEFAULT NULL::character varying,
    new_value character varying(191) DEFAULT NULL::character varying,
    note text,
    ip_address character varying(45) DEFAULT NULL::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.kyc_audit_logs OWNER TO smsagribd_sc;

--
-- Name: kyc_audit_logs_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.kyc_audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.kyc_audit_logs_id_seq OWNER TO smsagribd_sc;

--
-- Name: kyc_audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.kyc_audit_logs_id_seq OWNED BY smsagribd_sc.kyc_audit_logs.id;


--
-- Name: kyc_role_permissions; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.kyc_role_permissions (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    permissions text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.kyc_role_permissions OWNER TO smsagribd_sc;

--
-- Name: kyc_role_permissions_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.kyc_role_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.kyc_role_permissions_id_seq OWNER TO smsagribd_sc;

--
-- Name: kyc_role_permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.kyc_role_permissions_id_seq OWNED BY smsagribd_sc.kyc_role_permissions.id;


--
-- Name: messages; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.messages (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    batch_id numeric,
    sender_id numeric,
    gateway_id numeric,
    "to" character varying(32) NOT NULL,
    body text NOT NULL,
    is_unicode boolean DEFAULT false NOT NULL,
    parts bigint DEFAULT '1'::bigint NOT NULL,
    rate numeric(10,4) DEFAULT 0.0000 NOT NULL,
    cost numeric(10,4) DEFAULT 0.0000 NOT NULL,
    operator_name character varying(64) DEFAULT NULL::character varying,
    operator_cost numeric(10,4) DEFAULT 0.0000 NOT NULL,
    primary_dipping_cost numeric(10,4) DEFAULT 0.0000 NOT NULL,
    third_party_dipping_cost numeric(10,4) DEFAULT 0.0000 NOT NULL,
    tariff_name character varying(100) DEFAULT NULL::character varying,
    source character varying(64) DEFAULT 'portal'::character varying NOT NULL,
    source_ip character varying(64) DEFAULT NULL::character varying,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    delivery_status character varying(64) DEFAULT NULL::character varying,
    gateway_message_id character varying(255) DEFAULT NULL::character varying,
    poll_attempt bigint DEFAULT '0'::bigint NOT NULL,
    next_poll_at timestamp with time zone,
    sent_at timestamp with time zone,
    delivered_at timestamp with time zone,
    failed_at timestamp with time zone,
    failed_reason text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.messages OWNER TO smsagribd_sc;

--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.messages_id_seq OWNER TO smsagribd_sc;

--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.messages_id_seq OWNED BY smsagribd_sc.messages.id;


--
-- Name: operators; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.operators (
    id bigint NOT NULL,
    country_id numeric NOT NULL,
    name character varying(100) NOT NULL,
    codes text NOT NULL,
    masking_price numeric(10,4) DEFAULT 0.0000 NOT NULL,
    non_masking_price numeric(10,4) DEFAULT 0.0000 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.operators OWNER TO smsagribd_sc;

--
-- Name: operators_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.operators_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.operators_id_seq OWNER TO smsagribd_sc;

--
-- Name: operators_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.operators_id_seq OWNED BY smsagribd_sc.operators.id;


--
-- Name: password_reset_tokens; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.password_reset_tokens (
    email character varying(255) NOT NULL,
    token character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE smsagribd_sc.password_reset_tokens OWNER TO smsagribd_sc;

--
-- Name: payment_webhooks; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.payment_webhooks (
    id bigint NOT NULL,
    gateway character varying(64) NOT NULL,
    event_id character varying(255) NOT NULL,
    event_type character varying(64) DEFAULT 'payment.updated'::character varying NOT NULL,
    status character varying(64) DEFAULT 'pending'::character varying NOT NULL,
    payload text,
    headers text,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.payment_webhooks OWNER TO smsagribd_sc;

--
-- Name: payment_webhooks_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.payment_webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.payment_webhooks_id_seq OWNER TO smsagribd_sc;

--
-- Name: payment_webhooks_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.payment_webhooks_id_seq OWNED BY smsagribd_sc.payment_webhooks.id;


--
-- Name: phone_photo_sessions; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.phone_photo_sessions (
    id character varying(64) NOT NULL,
    session_token character varying(255) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    photo_data text,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.phone_photo_sessions OWNER TO smsagribd_sc;

--
-- Name: route_plan_gateways; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.route_plan_gateways (
    id bigint NOT NULL,
    route_plan_id numeric NOT NULL,
    gateway_id numeric NOT NULL,
    priority bigint DEFAULT '1'::bigint NOT NULL,
    weight bigint DEFAULT '100'::bigint NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.route_plan_gateways OWNER TO smsagribd_sc;

--
-- Name: route_plan_gateways_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.route_plan_gateways_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.route_plan_gateways_id_seq OWNER TO smsagribd_sc;

--
-- Name: route_plan_gateways_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.route_plan_gateways_id_seq OWNED BY smsagribd_sc.route_plan_gateways.id;


--
-- Name: route_plans; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.route_plans (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    code character varying(64) NOT NULL,
    description text,
    traffic_type character varying(64) DEFAULT 'transactional'::character varying NOT NULL,
    failover_strategy character varying(64) DEFAULT 'priority'::character varying NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    is_enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.route_plans OWNER TO smsagribd_sc;

--
-- Name: route_plans_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.route_plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.route_plans_id_seq OWNER TO smsagribd_sc;

--
-- Name: route_plans_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.route_plans_id_seq OWNED BY smsagribd_sc.route_plans.id;


--
-- Name: route_rules; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.route_rules (
    id bigint NOT NULL,
    route_plan_id numeric NOT NULL,
    name character varying(191) NOT NULL,
    is_enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.route_rules OWNER TO smsagribd_sc;

--
-- Name: route_rules_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.route_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.route_rules_id_seq OWNER TO smsagribd_sc;

--
-- Name: route_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.route_rules_id_seq OWNED BY smsagribd_sc.route_rules.id;


--
-- Name: sender_user; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.sender_user (
    id bigint NOT NULL,
    sender_id numeric NOT NULL,
    user_id numeric NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.sender_user OWNER TO smsagribd_sc;

--
-- Name: sender_user_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.sender_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.sender_user_id_seq OWNER TO smsagribd_sc;

--
-- Name: sender_user_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.sender_user_id_seq OWNED BY smsagribd_sc.sender_user.id;


--
-- Name: senders; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.senders (
    id bigint NOT NULL,
    user_id numeric,
    requested_by numeric,
    gateway_id numeric,
    sender_id character varying(100) NOT NULL,
    type character varying(64) DEFAULT 'shared'::character varying NOT NULL,
    approval_status character varying(32) DEFAULT 'approved'::character varying NOT NULL,
    is_enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.senders OWNER TO smsagribd_sc;

--
-- Name: senders_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.senders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.senders_id_seq OWNER TO smsagribd_sc;

--
-- Name: senders_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.senders_id_seq OWNED BY smsagribd_sc.senders.id;


--
-- Name: settings; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.settings (
    id bigint NOT NULL,
    "group" character varying(100) NOT NULL,
    name character varying(100) NOT NULL,
    payload text,
    locked boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.settings OWNER TO smsagribd_sc;

--
-- Name: settings_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.settings_id_seq OWNER TO smsagribd_sc;

--
-- Name: settings_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.settings_id_seq OWNED BY smsagribd_sc.settings.id;


--
-- Name: signup_slug_aliases; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.signup_slug_aliases (
    id bigint NOT NULL,
    tariff_plan_id numeric NOT NULL,
    slug character varying(191) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.signup_slug_aliases OWNER TO smsagribd_sc;

--
-- Name: signup_slug_aliases_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.signup_slug_aliases_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.signup_slug_aliases_id_seq OWNER TO smsagribd_sc;

--
-- Name: signup_slug_aliases_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.signup_slug_aliases_id_seq OWNED BY smsagribd_sc.signup_slug_aliases.id;


--
-- Name: sms_logs; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.sms_logs (
    id bigint NOT NULL,
    phone character varying(32) DEFAULT NULL::character varying,
    gateway_name character varying(191) DEFAULT NULL::character varying,
    event character varying(100) NOT NULL,
    error_message text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.sms_logs OWNER TO smsagribd_sc;

--
-- Name: sms_logs_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.sms_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.sms_logs_id_seq OWNER TO smsagribd_sc;

--
-- Name: sms_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.sms_logs_id_seq OWNED BY smsagribd_sc.sms_logs.id;


--
-- Name: sp_distributions; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.sp_distributions (
    id bigint NOT NULL,
    sp_id numeric NOT NULL,
    customer_id numeric NOT NULL,
    amount numeric(15,4) DEFAULT 0.0000 NOT NULL,
    note text,
    distributed_by numeric NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.sp_distributions OWNER TO smsagribd_sc;

--
-- Name: sp_distributions_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.sp_distributions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.sp_distributions_id_seq OWNER TO smsagribd_sc;

--
-- Name: sp_distributions_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.sp_distributions_id_seq OWNED BY smsagribd_sc.sp_distributions.id;


--
-- Name: tariff_plan_assignments; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.tariff_plan_assignments (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    tariff_plan_id numeric NOT NULL,
    assigned_by numeric,
    notes text,
    status character varying(32) DEFAULT 'active'::character varying NOT NULL,
    effective_from timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.tariff_plan_assignments OWNER TO smsagribd_sc;

--
-- Name: tariff_plan_assignments_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.tariff_plan_assignments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.tariff_plan_assignments_id_seq OWNER TO smsagribd_sc;

--
-- Name: tariff_plan_assignments_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.tariff_plan_assignments_id_seq OWNED BY smsagribd_sc.tariff_plan_assignments.id;


--
-- Name: tariff_plans; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.tariff_plans (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    type character varying(50) DEFAULT 'standard'::character varying NOT NULL,
    description text,
    currency character varying(10) DEFAULT 'BDT'::character varying NOT NULL,
    status character varying(32) DEFAULT 'active'::character varying NOT NULL,
    created_by numeric,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.tariff_plans OWNER TO smsagribd_sc;

--
-- Name: tariff_plans_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.tariff_plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.tariff_plans_id_seq OWNER TO smsagribd_sc;

--
-- Name: tariff_plans_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.tariff_plans_id_seq OWNED BY smsagribd_sc.tariff_plans.id;


--
-- Name: tariff_rate_history; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.tariff_rate_history (
    id bigint NOT NULL,
    tariff_rate_id numeric,
    tariff_plan_id numeric NOT NULL,
    operator_id numeric NOT NULL,
    event character varying(64) DEFAULT 'update'::character varying NOT NULL,
    old_masking_price numeric(10,4) DEFAULT NULL::numeric,
    new_masking_price numeric(10,4) DEFAULT NULL::numeric,
    old_non_masking_price numeric(10,4) DEFAULT NULL::numeric,
    new_non_masking_price numeric(10,4) DEFAULT NULL::numeric,
    changed_by numeric,
    changed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.tariff_rate_history OWNER TO smsagribd_sc;

--
-- Name: tariff_rate_history_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.tariff_rate_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.tariff_rate_history_id_seq OWNER TO smsagribd_sc;

--
-- Name: tariff_rate_history_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.tariff_rate_history_id_seq OWNED BY smsagribd_sc.tariff_rate_history.id;


--
-- Name: tariff_rates; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.tariff_rates (
    id bigint NOT NULL,
    tariff_plan_id numeric NOT NULL,
    operator_id numeric NOT NULL,
    masking_price numeric(10,4) DEFAULT 0.0000 NOT NULL,
    non_masking_price numeric(10,4) DEFAULT 0.0000 NOT NULL,
    cost numeric(10,4) DEFAULT 0.0000 NOT NULL,
    notes text,
    effective_from timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.tariff_rates OWNER TO smsagribd_sc;

--
-- Name: tariff_rates_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.tariff_rates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.tariff_rates_id_seq OWNER TO smsagribd_sc;

--
-- Name: tariff_rates_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.tariff_rates_id_seq OWNED BY smsagribd_sc.tariff_rates.id;


--
-- Name: templates; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.templates (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    name character varying(191) NOT NULL,
    body text NOT NULL,
    variables text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.templates OWNER TO smsagribd_sc;

--
-- Name: templates_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.templates_id_seq OWNER TO smsagribd_sc;

--
-- Name: templates_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.templates_id_seq OWNED BY smsagribd_sc.templates.id;


--
-- Name: tickets; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.tickets (
    id bigint NOT NULL,
    user_id numeric DEFAULT '0'::numeric NOT NULL,
    subject character varying(255) DEFAULT ''::character varying NOT NULL,
    status character varying(50) DEFAULT 'open'::character varying NOT NULL,
    priority character varying(50) DEFAULT 'normal'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.tickets OWNER TO smsagribd_sc;

--
-- Name: tickets_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.tickets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.tickets_id_seq OWNER TO smsagribd_sc;

--
-- Name: tickets_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.tickets_id_seq OWNED BY smsagribd_sc.tickets.id;


--
-- Name: transactions; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.transactions (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    transactionable_type character varying(100) DEFAULT ''::character varying NOT NULL,
    transactionable_id numeric DEFAULT '0'::numeric NOT NULL,
    direction character varying(16) DEFAULT 'credit'::character varying NOT NULL,
    amount numeric(15,4) DEFAULT 0.0000 NOT NULL,
    balance_before numeric(15,4) DEFAULT 0.0000 NOT NULL,
    balance_after numeric(15,4) DEFAULT 0.0000 NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.transactions OWNER TO smsagribd_sc;

--
-- Name: transactions_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.transactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.transactions_id_seq OWNER TO smsagribd_sc;

--
-- Name: transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.transactions_id_seq OWNED BY smsagribd_sc.transactions.id;


--
-- Name: user_operator_rates; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.user_operator_rates (
    id bigint NOT NULL,
    user_id numeric NOT NULL,
    operator_id numeric NOT NULL,
    masking_price numeric(10,4) DEFAULT 0.0000 NOT NULL,
    non_masking_price numeric(10,4) DEFAULT 0.0000 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.user_operator_rates OWNER TO smsagribd_sc;

--
-- Name: user_operator_rates_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.user_operator_rates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.user_operator_rates_id_seq OWNER TO smsagribd_sc;

--
-- Name: user_operator_rates_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.user_operator_rates_id_seq OWNED BY smsagribd_sc.user_operator_rates.id;


--
-- Name: users; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.users (
    id bigint NOT NULL,
    sp_id numeric,
    name character varying(255) DEFAULT ''::character varying NOT NULL,
    email character varying(255) NOT NULL,
    phone character varying(32) DEFAULT NULL::character varying,
    password character varying(255) NOT NULL,
    balance numeric(15,4) DEFAULT 0.0000 NOT NULL,
    type character varying(32) DEFAULT 'customer'::character varying NOT NULL,
    status character varying(32) DEFAULT 'approved'::character varying NOT NULL,
    kyc_status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    kyc_submitted_at timestamp with time zone,
    kyc_verified_at timestamp with time zone,
    kyc_resubmit_fields text,
    kyc_rejected_reason text,
    nid_number character varying(64) DEFAULT NULL::character varying,
    front_nid character varying(255) DEFAULT NULL::character varying,
    back_nid character varying(255) DEFAULT NULL::character varying,
    live_photo character varying(255) DEFAULT NULL::character varying,
    business_name character varying(255) DEFAULT NULL::character varying,
    company_name character varying(255) DEFAULT NULL::character varying,
    business_address text,
    trade_license_number character varying(64) DEFAULT NULL::character varying,
    trade_license_file character varying(255) DEFAULT NULL::character varying,
    is_admin_created boolean DEFAULT false NOT NULL,
    created_by_admin boolean DEFAULT false NOT NULL,
    sms_auto_approved boolean DEFAULT true NOT NULL,
    api_access boolean DEFAULT true NOT NULL,
    last_login_at timestamp with time zone,
    remember_token character varying(100) DEFAULT NULL::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.users OWNER TO smsagribd_sc;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.users_id_seq OWNER TO smsagribd_sc;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.users_id_seq OWNED BY smsagribd_sc.users.id;


--
-- Name: verification_otps; Type: TABLE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TABLE smsagribd_sc.verification_otps (
    id bigint NOT NULL,
    identifier character varying(191) NOT NULL,
    type character varying(32) DEFAULT 'phone'::character varying NOT NULL,
    otp_hash character varying(255) NOT NULL,
    attempts bigint DEFAULT '0'::bigint NOT NULL,
    last_sent_at timestamp with time zone,
    expires_at timestamp with time zone NOT NULL,
    verified_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);


ALTER TABLE smsagribd_sc.verification_otps OWNER TO smsagribd_sc;

--
-- Name: verification_otps_id_seq; Type: SEQUENCE; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE SEQUENCE smsagribd_sc.verification_otps_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE smsagribd_sc.verification_otps_id_seq OWNER TO smsagribd_sc;

--
-- Name: verification_otps_id_seq; Type: SEQUENCE OWNED BY; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER SEQUENCE smsagribd_sc.verification_otps_id_seq OWNED BY smsagribd_sc.verification_otps.id;


--
-- Name: access_tokens id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.access_tokens ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.access_tokens_id_seq'::regclass);


--
-- Name: batches id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.batches ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.batches_id_seq'::regclass);


--
-- Name: campaigns id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.campaigns ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.campaigns_id_seq'::regclass);


--
-- Name: contact_group id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.contact_group ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.contact_group_id_seq'::regclass);


--
-- Name: contacts id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.contacts ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.contacts_id_seq'::regclass);


--
-- Name: countries id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.countries ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.countries_id_seq'::regclass);


--
-- Name: deposits id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.deposits ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.deposits_id_seq'::regclass);


--
-- Name: gateway_operator_rates id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.gateway_operator_rates ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.gateway_operator_rates_id_seq'::regclass);


--
-- Name: gateways id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.gateways ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.gateways_id_seq'::regclass);


--
-- Name: groups id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.groups ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.groups_id_seq'::regclass);


--
-- Name: ip_whitelist_audit_logs id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.ip_whitelist_audit_logs ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.ip_whitelist_audit_logs_id_seq'::regclass);


--
-- Name: ip_whitelist_requests id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.ip_whitelist_requests ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.ip_whitelist_requests_id_seq'::regclass);


--
-- Name: kyc_audit_logs id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.kyc_audit_logs ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.kyc_audit_logs_id_seq'::regclass);


--
-- Name: kyc_role_permissions id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.kyc_role_permissions ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.kyc_role_permissions_id_seq'::regclass);


--
-- Name: messages id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.messages ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.messages_id_seq'::regclass);


--
-- Name: operators id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.operators ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.operators_id_seq'::regclass);


--
-- Name: payment_webhooks id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.payment_webhooks ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.payment_webhooks_id_seq'::regclass);


--
-- Name: route_plan_gateways id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.route_plan_gateways ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.route_plan_gateways_id_seq'::regclass);


--
-- Name: route_plans id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.route_plans ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.route_plans_id_seq'::regclass);


--
-- Name: route_rules id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.route_rules ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.route_rules_id_seq'::regclass);


--
-- Name: sender_user id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.sender_user ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.sender_user_id_seq'::regclass);


--
-- Name: senders id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.senders ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.senders_id_seq'::regclass);


--
-- Name: settings id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.settings ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.settings_id_seq'::regclass);


--
-- Name: signup_slug_aliases id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.signup_slug_aliases ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.signup_slug_aliases_id_seq'::regclass);


--
-- Name: sms_logs id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.sms_logs ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.sms_logs_id_seq'::regclass);


--
-- Name: sp_distributions id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.sp_distributions ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.sp_distributions_id_seq'::regclass);


--
-- Name: tariff_plan_assignments id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tariff_plan_assignments ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.tariff_plan_assignments_id_seq'::regclass);


--
-- Name: tariff_plans id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tariff_plans ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.tariff_plans_id_seq'::regclass);


--
-- Name: tariff_rate_history id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tariff_rate_history ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.tariff_rate_history_id_seq'::regclass);


--
-- Name: tariff_rates id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tariff_rates ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.tariff_rates_id_seq'::regclass);


--
-- Name: templates id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.templates ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.templates_id_seq'::regclass);


--
-- Name: tickets id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tickets ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.tickets_id_seq'::regclass);


--
-- Name: transactions id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.transactions ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.transactions_id_seq'::regclass);


--
-- Name: user_operator_rates id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.user_operator_rates ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.user_operator_rates_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.users ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.users_id_seq'::regclass);


--
-- Name: verification_otps id; Type: DEFAULT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.verification_otps ALTER COLUMN id SET DEFAULT nextval('smsagribd_sc.verification_otps_id_seq'::regclass);


--
-- Name: access_tokens idx_17980_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.access_tokens
    ADD CONSTRAINT idx_17980_primary PRIMARY KEY (id);


--
-- Name: batches idx_17993_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.batches
    ADD CONSTRAINT idx_17993_primary PRIMARY KEY (id);


--
-- Name: cache idx_18019_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.cache
    ADD CONSTRAINT idx_18019_primary PRIMARY KEY (key);


--
-- Name: campaigns idx_18028_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.campaigns
    ADD CONSTRAINT idx_18028_primary PRIMARY KEY (id);


--
-- Name: contacts idx_18042_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.contacts
    ADD CONSTRAINT idx_18042_primary PRIMARY KEY (id);


--
-- Name: contact_group idx_18056_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.contact_group
    ADD CONSTRAINT idx_18056_primary PRIMARY KEY (id);


--
-- Name: countries idx_18067_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.countries
    ADD CONSTRAINT idx_18067_primary PRIMARY KEY (id);


--
-- Name: deposits idx_18079_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.deposits
    ADD CONSTRAINT idx_18079_primary PRIMARY KEY (id);


--
-- Name: gateways idx_18098_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.gateways
    ADD CONSTRAINT idx_18098_primary PRIMARY KEY (id);


--
-- Name: gateway_operator_rates idx_18134_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.gateway_operator_rates
    ADD CONSTRAINT idx_18134_primary PRIMARY KEY (id);


--
-- Name: groups idx_18151_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.groups
    ADD CONSTRAINT idx_18151_primary PRIMARY KEY (id);


--
-- Name: ip_whitelist_audit_logs idx_18162_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.ip_whitelist_audit_logs
    ADD CONSTRAINT idx_18162_primary PRIMARY KEY (id);


--
-- Name: ip_whitelist_requests idx_18179_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.ip_whitelist_requests
    ADD CONSTRAINT idx_18179_primary PRIMARY KEY (id);


--
-- Name: kyc_audit_logs idx_18193_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.kyc_audit_logs
    ADD CONSTRAINT idx_18193_primary PRIMARY KEY (id);


--
-- Name: kyc_role_permissions idx_18209_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.kyc_role_permissions
    ADD CONSTRAINT idx_18209_primary PRIMARY KEY (id);


--
-- Name: messages idx_18219_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.messages
    ADD CONSTRAINT idx_18219_primary PRIMARY KEY (id);


--
-- Name: operators idx_18256_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.operators
    ADD CONSTRAINT idx_18256_primary PRIMARY KEY (id);


--
-- Name: password_reset_tokens idx_18273_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.password_reset_tokens
    ADD CONSTRAINT idx_18273_primary PRIMARY KEY (email);


--
-- Name: payment_webhooks idx_18282_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.payment_webhooks
    ADD CONSTRAINT idx_18282_primary PRIMARY KEY (id);


--
-- Name: phone_photo_sessions idx_18296_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.phone_photo_sessions
    ADD CONSTRAINT idx_18296_primary PRIMARY KEY (id);


--
-- Name: route_plans idx_18308_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.route_plans
    ADD CONSTRAINT idx_18308_primary PRIMARY KEY (id);


--
-- Name: route_plan_gateways idx_18327_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.route_plan_gateways
    ADD CONSTRAINT idx_18327_primary PRIMARY KEY (id);


--
-- Name: route_rules idx_18344_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.route_rules
    ADD CONSTRAINT idx_18344_primary PRIMARY KEY (id);


--
-- Name: senders idx_18357_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.senders
    ADD CONSTRAINT idx_18357_primary PRIMARY KEY (id);


--
-- Name: sender_user idx_18373_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.sender_user
    ADD CONSTRAINT idx_18373_primary PRIMARY KEY (id);


--
-- Name: settings idx_18384_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.settings
    ADD CONSTRAINT idx_18384_primary PRIMARY KEY (id);


--
-- Name: signup_slug_aliases idx_18397_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.signup_slug_aliases
    ADD CONSTRAINT idx_18397_primary PRIMARY KEY (id);


--
-- Name: sms_logs idx_18408_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.sms_logs
    ADD CONSTRAINT idx_18408_primary PRIMARY KEY (id);


--
-- Name: sp_distributions idx_18420_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.sp_distributions
    ADD CONSTRAINT idx_18420_primary PRIMARY KEY (id);


--
-- Name: tariff_plans idx_18434_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tariff_plans
    ADD CONSTRAINT idx_18434_primary PRIMARY KEY (id);


--
-- Name: tariff_plan_assignments idx_18451_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tariff_plan_assignments
    ADD CONSTRAINT idx_18451_primary PRIMARY KEY (id);


--
-- Name: tariff_rates idx_18466_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tariff_rates
    ADD CONSTRAINT idx_18466_primary PRIMARY KEY (id);


--
-- Name: tariff_rate_history idx_18485_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tariff_rate_history
    ADD CONSTRAINT idx_18485_primary PRIMARY KEY (id);


--
-- Name: templates idx_18504_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.templates
    ADD CONSTRAINT idx_18504_primary PRIMARY KEY (id);


--
-- Name: tickets idx_18516_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.tickets
    ADD CONSTRAINT idx_18516_primary PRIMARY KEY (id);


--
-- Name: transactions idx_18533_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.transactions
    ADD CONSTRAINT idx_18533_primary PRIMARY KEY (id);


--
-- Name: users idx_18555_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.users
    ADD CONSTRAINT idx_18555_primary PRIMARY KEY (id);


--
-- Name: user_operator_rates idx_18594_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.user_operator_rates
    ADD CONSTRAINT idx_18594_primary PRIMARY KEY (id);


--
-- Name: verification_otps idx_18609_primary; Type: CONSTRAINT; Schema: smsagribd_sc; Owner: smsagribd_sc
--

ALTER TABLE ONLY smsagribd_sc.verification_otps
    ADD CONSTRAINT idx_18609_primary PRIMARY KEY (id);


--
-- Name: idx_17980_idx_access_tokens_tokenable; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_17980_idx_access_tokens_tokenable ON smsagribd_sc.access_tokens USING btree (tokenable_type, tokenable_id);


--
-- Name: idx_17980_uk_access_tokens_token; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE UNIQUE INDEX idx_17980_uk_access_tokens_token ON smsagribd_sc.access_tokens USING btree (token);


--
-- Name: idx_17993_idx_batches_status; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_17993_idx_batches_status ON smsagribd_sc.batches USING btree (status);


--
-- Name: idx_17993_idx_batches_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_17993_idx_batches_user ON smsagribd_sc.batches USING btree (user_id);


--
-- Name: idx_18019_idx_cache_expiration; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18019_idx_cache_expiration ON smsagribd_sc.cache USING btree (expiration);


--
-- Name: idx_18042_idx_contacts_phone; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18042_idx_contacts_phone ON smsagribd_sc.contacts USING btree (phone);


--
-- Name: idx_18042_idx_contacts_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18042_idx_contacts_user ON smsagribd_sc.contacts USING btree (user_id);


--
-- Name: idx_18056_uk_contact_group; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE UNIQUE INDEX idx_18056_uk_contact_group ON smsagribd_sc.contact_group USING btree (contact_id, group_id);


--
-- Name: idx_18079_idx_deposits_status; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18079_idx_deposits_status ON smsagribd_sc.deposits USING btree (status);


--
-- Name: idx_18079_idx_deposits_txn; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18079_idx_deposits_txn ON smsagribd_sc.deposits USING btree (transaction_id);


--
-- Name: idx_18079_idx_deposits_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18079_idx_deposits_user ON smsagribd_sc.deposits USING btree (user_id);


--
-- Name: idx_18134_idx_gor_gateway; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18134_idx_gor_gateway ON smsagribd_sc.gateway_operator_rates USING btree (gateway_id);


--
-- Name: idx_18134_idx_gor_operator; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18134_idx_gor_operator ON smsagribd_sc.gateway_operator_rates USING btree (operator_id);


--
-- Name: idx_18151_idx_groups_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18151_idx_groups_user ON smsagribd_sc.groups USING btree (user_id);


--
-- Name: idx_18162_idx_ipal_request; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18162_idx_ipal_request ON smsagribd_sc.ip_whitelist_audit_logs USING btree (ip_whitelist_request_id);


--
-- Name: idx_18162_idx_ipal_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18162_idx_ipal_user ON smsagribd_sc.ip_whitelist_audit_logs USING btree (user_id);


--
-- Name: idx_18179_idx_ip_status; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18179_idx_ip_status ON smsagribd_sc.ip_whitelist_requests USING btree (status);


--
-- Name: idx_18179_idx_ip_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18179_idx_ip_user ON smsagribd_sc.ip_whitelist_requests USING btree (user_id);


--
-- Name: idx_18193_idx_kyc_audit_admin; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18193_idx_kyc_audit_admin ON smsagribd_sc.kyc_audit_logs USING btree (admin_id);


--
-- Name: idx_18193_idx_kyc_audit_created; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18193_idx_kyc_audit_created ON smsagribd_sc.kyc_audit_logs USING btree (created_at);


--
-- Name: idx_18193_idx_kyc_audit_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18193_idx_kyc_audit_user ON smsagribd_sc.kyc_audit_logs USING btree (user_id);


--
-- Name: idx_18209_idx_kyc_role_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18209_idx_kyc_role_user ON smsagribd_sc.kyc_role_permissions USING btree (user_id);


--
-- Name: idx_18209_user_id; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE UNIQUE INDEX idx_18209_user_id ON smsagribd_sc.kyc_role_permissions USING btree (user_id);


--
-- Name: idx_18219_idx_messages_batch; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18219_idx_messages_batch ON smsagribd_sc.messages USING btree (batch_id);


--
-- Name: idx_18219_idx_messages_gateway_msg; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18219_idx_messages_gateway_msg ON smsagribd_sc.messages USING btree (gateway_message_id);


--
-- Name: idx_18219_idx_messages_next_poll; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18219_idx_messages_next_poll ON smsagribd_sc.messages USING btree (next_poll_at);


--
-- Name: idx_18219_idx_messages_status; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18219_idx_messages_status ON smsagribd_sc.messages USING btree (status);


--
-- Name: idx_18219_idx_messages_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18219_idx_messages_user ON smsagribd_sc.messages USING btree (user_id);


--
-- Name: idx_18256_idx_operators_country; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18256_idx_operators_country ON smsagribd_sc.operators USING btree (country_id);


--
-- Name: idx_18282_idx_payment_webhooks_status; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18282_idx_payment_webhooks_status ON smsagribd_sc.payment_webhooks USING btree (status);


--
-- Name: idx_18282_uk_payment_webhooks; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE UNIQUE INDEX idx_18282_uk_payment_webhooks ON smsagribd_sc.payment_webhooks USING btree (gateway, event_id);


--
-- Name: idx_18296_idx_pps_token; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18296_idx_pps_token ON smsagribd_sc.phone_photo_sessions USING btree (session_token);


--
-- Name: idx_18308_code; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE UNIQUE INDEX idx_18308_code ON smsagribd_sc.route_plans USING btree (code);


--
-- Name: idx_18327_idx_rpg_gateway; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18327_idx_rpg_gateway ON smsagribd_sc.route_plan_gateways USING btree (gateway_id);


--
-- Name: idx_18327_idx_rpg_plan; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18327_idx_rpg_plan ON smsagribd_sc.route_plan_gateways USING btree (route_plan_id);


--
-- Name: idx_18357_idx_senders_gateway_id; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18357_idx_senders_gateway_id ON smsagribd_sc.senders USING btree (gateway_id);


--
-- Name: idx_18357_idx_senders_sender_id; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18357_idx_senders_sender_id ON smsagribd_sc.senders USING btree (sender_id);


--
-- Name: idx_18373_idx_su_sender; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18373_idx_su_sender ON smsagribd_sc.sender_user USING btree (sender_id);


--
-- Name: idx_18373_idx_su_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18373_idx_su_user ON smsagribd_sc.sender_user USING btree (user_id);


--
-- Name: idx_18384_uk_settings_group_name; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE UNIQUE INDEX idx_18384_uk_settings_group_name ON smsagribd_sc.settings USING btree ("group", name);


--
-- Name: idx_18397_slug; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE UNIQUE INDEX idx_18397_slug ON smsagribd_sc.signup_slug_aliases USING btree (slug);


--
-- Name: idx_18408_idx_sms_logs_created; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18408_idx_sms_logs_created ON smsagribd_sc.sms_logs USING btree (created_at);


--
-- Name: idx_18408_idx_sms_logs_event; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18408_idx_sms_logs_event ON smsagribd_sc.sms_logs USING btree (event);


--
-- Name: idx_18420_idx_spd_customer; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18420_idx_spd_customer ON smsagribd_sc.sp_distributions USING btree (customer_id);


--
-- Name: idx_18420_idx_spd_sp; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18420_idx_spd_sp ON smsagribd_sc.sp_distributions USING btree (sp_id);


--
-- Name: idx_18434_slug; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE UNIQUE INDEX idx_18434_slug ON smsagribd_sc.tariff_plans USING btree (slug);


--
-- Name: idx_18451_idx_tpa_plan; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18451_idx_tpa_plan ON smsagribd_sc.tariff_plan_assignments USING btree (tariff_plan_id);


--
-- Name: idx_18451_idx_tpa_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18451_idx_tpa_user ON smsagribd_sc.tariff_plan_assignments USING btree (user_id);


--
-- Name: idx_18466_idx_tr_operator; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18466_idx_tr_operator ON smsagribd_sc.tariff_rates USING btree (operator_id);


--
-- Name: idx_18466_idx_tr_plan; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18466_idx_tr_plan ON smsagribd_sc.tariff_rates USING btree (tariff_plan_id);


--
-- Name: idx_18504_idx_templates_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18504_idx_templates_user ON smsagribd_sc.templates USING btree (user_id);


--
-- Name: idx_18533_idx_transactions_type_id; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18533_idx_transactions_type_id ON smsagribd_sc.transactions USING btree (transactionable_type, transactionable_id);


--
-- Name: idx_18533_idx_transactions_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18533_idx_transactions_user ON smsagribd_sc.transactions USING btree (user_id);


--
-- Name: idx_18555_idx_users_email; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18555_idx_users_email ON smsagribd_sc.users USING btree (email);


--
-- Name: idx_18555_idx_users_phone; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18555_idx_users_phone ON smsagribd_sc.users USING btree (phone);


--
-- Name: idx_18555_idx_users_sp_id; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18555_idx_users_sp_id ON smsagribd_sc.users USING btree (sp_id);


--
-- Name: idx_18555_idx_users_status; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18555_idx_users_status ON smsagribd_sc.users USING btree (status);


--
-- Name: idx_18555_idx_users_type; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18555_idx_users_type ON smsagribd_sc.users USING btree (type);


--
-- Name: idx_18594_idx_uor_operator; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18594_idx_uor_operator ON smsagribd_sc.user_operator_rates USING btree (operator_id);


--
-- Name: idx_18594_idx_uor_user; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18594_idx_uor_user ON smsagribd_sc.user_operator_rates USING btree (user_id);


--
-- Name: idx_18609_idx_votp_ident_type; Type: INDEX; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE INDEX idx_18609_idx_votp_ident_type ON smsagribd_sc.verification_otps USING btree (identifier, type);


--
-- Name: access_tokens on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.access_tokens FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_access_tokens();


--
-- Name: batches on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.batches FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_batches();


--
-- Name: campaigns on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.campaigns FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_campaigns();


--
-- Name: contact_group on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.contact_group FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_contact_group();


--
-- Name: contacts on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.contacts FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_contacts();


--
-- Name: countries on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.countries FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_countries();


--
-- Name: deposits on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.deposits FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_deposits();


--
-- Name: gateway_operator_rates on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.gateway_operator_rates FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_gateway_operator_rates();


--
-- Name: gateways on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.gateways FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_gateways();


--
-- Name: groups on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.groups FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_groups();


--
-- Name: ip_whitelist_audit_logs on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.ip_whitelist_audit_logs FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_ip_whitelist_audit_logs();


--
-- Name: ip_whitelist_requests on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.ip_whitelist_requests FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_ip_whitelist_requests();


--
-- Name: kyc_audit_logs on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.kyc_audit_logs FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_kyc_audit_logs();


--
-- Name: kyc_role_permissions on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.kyc_role_permissions FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_kyc_role_permissions();


--
-- Name: messages on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.messages FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_messages();


--
-- Name: operators on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.operators FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_operators();


--
-- Name: payment_webhooks on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.payment_webhooks FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_payment_webhooks();


--
-- Name: phone_photo_sessions on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.phone_photo_sessions FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_phone_photo_sessions();


--
-- Name: route_plan_gateways on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.route_plan_gateways FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_route_plan_gateways();


--
-- Name: route_plans on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.route_plans FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_route_plans();


--
-- Name: route_rules on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.route_rules FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_route_rules();


--
-- Name: sender_user on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.sender_user FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_sender_user();


--
-- Name: senders on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.senders FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_senders();


--
-- Name: settings on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.settings FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_settings();


--
-- Name: signup_slug_aliases on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.signup_slug_aliases FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_signup_slug_aliases();


--
-- Name: sms_logs on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.sms_logs FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_sms_logs();


--
-- Name: sp_distributions on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.sp_distributions FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_sp_distributions();


--
-- Name: tariff_plan_assignments on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.tariff_plan_assignments FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_plan_assignments();


--
-- Name: tariff_plans on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.tariff_plans FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_plans();


--
-- Name: tariff_rate_history on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.tariff_rate_history FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_rate_history();


--
-- Name: tariff_rates on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.tariff_rates FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_tariff_rates();


--
-- Name: templates on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.templates FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_templates();


--
-- Name: tickets on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.tickets FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_tickets();


--
-- Name: transactions on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.transactions FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_transactions();


--
-- Name: user_operator_rates on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.user_operator_rates FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_user_operator_rates();


--
-- Name: users on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.users FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_users();


--
-- Name: verification_otps on_update_current_timestamp; Type: TRIGGER; Schema: smsagribd_sc; Owner: smsagribd_sc
--

CREATE TRIGGER on_update_current_timestamp BEFORE UPDATE ON smsagribd_sc.verification_otps FOR EACH ROW EXECUTE FUNCTION smsagribd_sc.on_update_current_timestamp_verification_otps();


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

GRANT ALL ON SCHEMA public TO smsagribd_sc;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON TABLES TO smsagribd_sc;


--
-- PostgreSQL database dump complete
--

\unrestrict R37nrIu0qgfQ5OMAnbJj5vogEayRzWKwMWr8BvpgOYwfueDxtrc3Jx5vMvs87u7


--
-- PostgreSQL database dump
--

\restrict dN5DqlqSrhOZrJoZVewEougEt67A3lnfoO7Gxb31ZKvODfn9SHEMHuUjQsbpzbQ

-- Dumped from database version 18.1
-- Dumped by pg_dump version 18.1

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
-- Name: collection; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA collection;


--
-- Name: common; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA common;


--
-- Name: revival; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA revival;


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: cheque_clearing_status; Type: TABLE; Schema: collection; Owner: -
--

CREATE TABLE collection.cheque_clearing_status (
    cheque_id character varying NOT NULL,
    payment_id character varying NOT NULL,
    request_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    cheque_number character varying NOT NULL,
    bank_name character varying NOT NULL,
    cheque_date date NOT NULL,
    amount numeric NOT NULL,
    clearance_status character varying NOT NULL,
    next_due_date date,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: collection_batch_tracking; Type: TABLE; Schema: collection; Owner: -
--

CREATE TABLE collection.collection_batch_tracking (
    batch_id character varying NOT NULL,
    request_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    premium_payment_id character varying NOT NULL,
    installment_payment_id character varying NOT NULL,
    collection_complete boolean DEFAULT false NOT NULL,
    collection_date date NOT NULL,
    combined_receipt_id character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    completed_at timestamp without time zone
);


--
-- Name: payment_receipts; Type: TABLE; Schema: collection; Owner: -
--

CREATE TABLE collection.payment_receipts (
    receipt_id character varying NOT NULL,
    receipt_number character varying NOT NULL,
    request_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    is_combined_receipt boolean DEFAULT false NOT NULL,
    payment_ids character varying[] NOT NULL,
    total_amount numeric NOT NULL,
    receipt_date date NOT NULL,
    payment_mode character varying NOT NULL,
    document_path text,
    generated_at timestamp without time zone DEFAULT now() NOT NULL,
    generated_by character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: payment_transactions; Type: TABLE; Schema: collection; Owner: -
--

CREATE TABLE collection.payment_transactions (
    payment_id character varying NOT NULL,
    request_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    collection_batch_id character varying NOT NULL,
    linked_payment_id character varying,
    cheque_id character varying,
    payment_type character varying NOT NULL,
    installment_number integer,
    amount numeric NOT NULL,
    tax_amount numeric NOT NULL,
    total_amount numeric NOT NULL,
    payment_mode character varying NOT NULL,
    payment_status character varying NOT NULL,
    collection_date date NOT NULL,
    payment_date timestamp without time zone,
    receipt_id character varying,
    tigerbeetle_transfer_id character varying,
    collected_by character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: suspense_accounts; Type: TABLE; Schema: collection; Owner: -
--

CREATE TABLE collection.suspense_accounts (
    suspense_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    request_id character varying,
    suspense_type character varying NOT NULL,
    amount numeric NOT NULL,
    is_reversed boolean DEFAULT false NOT NULL,
    reversal_date timestamp without time zone,
    reversal_authorized_by character varying,
    reversal_reason character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    created_by character varying,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    reason text,
    source_payment_ids text,
    tigerbeetle_transfer_id character varying,
    suspense_account_type character varying(20) DEFAULT 'REVIVAL_SUSPENSE'::character varying
);


--
-- Name: suspense_reversal_audit; Type: TABLE; Schema: collection; Owner: -
--

CREATE TABLE collection.suspense_reversal_audit (
    suspense_id character varying NOT NULL,
    reversed_at timestamp without time zone DEFAULT now() NOT NULL,
    authorized_by character varying NOT NULL,
    reversal_reason character varying NOT NULL,
    amount_reversed numeric NOT NULL,
    is_first_collection boolean NOT NULL,
    reversal_allowed boolean NOT NULL,
    rejection_reason character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: configuration_audit; Type: TABLE; Schema: common; Owner: -
--

CREATE TABLE common.configuration_audit (
    audit_id character varying NOT NULL,
    config_key character varying NOT NULL,
    old_value character varying NOT NULL,
    new_value character varying NOT NULL,
    changed_by character varying NOT NULL,
    changed_reason character varying,
    affected_policies_count integer,
    changed_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: policies; Type: TABLE; Schema: common; Owner: -
--

CREATE TABLE common.policies (
    policy_number character varying NOT NULL,
    customer_id character varying NOT NULL,
    customer_name character varying NOT NULL,
    product_code character varying NOT NULL,
    product_name character varying NOT NULL,
    policy_status character varying NOT NULL,
    premium_frequency character varying NOT NULL,
    premium_amount numeric NOT NULL,
    sum_assured numeric NOT NULL,
    paid_to_date date,
    maturity_date date NOT NULL,
    date_of_commencement date NOT NULL,
    billing_method character varying NOT NULL,
    revival_count integer DEFAULT 0 NOT NULL,
    last_revival_date date,
    office character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: system_configuration; Type: TABLE; Schema: common; Owner: -
--

CREATE TABLE common.system_configuration (
    config_key character varying NOT NULL,
    config_value character varying NOT NULL,
    is_configurable boolean DEFAULT false NOT NULL,
    description character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: tigerbeetle_accounts; Type: TABLE; Schema: common; Owner: -
--

CREATE TABLE common.tigerbeetle_accounts (
    account_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    premium_account_id character varying NOT NULL,
    revival_account_id character varying NOT NULL,
    loan_account_id character varying NOT NULL,
    combined_suspense_account_id character varying NOT NULL,
    revival_suspense_account_id character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: document_references; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.document_references (
    reference_id character varying NOT NULL,
    request_id character varying NOT NULL,
    document_type character varying NOT NULL,
    document_id character varying NOT NULL,
    document_name character varying,
    document_path text,
    received_date timestamp without time zone,
    received_by character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: generated_letters; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.generated_letters (
    letter_id character varying NOT NULL,
    request_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    letter_type character varying NOT NULL,
    document_path text,
    document_hash character varying,
    generated_at timestamp without time zone DEFAULT now() NOT NULL,
    generated_by character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: installment_schedules; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.installment_schedules (
    schedule_id character varying NOT NULL,
    request_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    installment_number integer NOT NULL,
    installment_amount numeric NOT NULL,
    tax_amount numeric NOT NULL,
    total_amount numeric NOT NULL,
    due_date date NOT NULL,
    payment_date timestamp without time zone,
    is_paid boolean DEFAULT false NOT NULL,
    grace_period_days integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_due_date_positive CHECK ((grace_period_days = 0))
);


--
-- Name: request_comments; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.request_comments (
    comment_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    request_id uuid NOT NULL,
    ticket_id character varying(20) NOT NULL,
    policy_number character varying(13) NOT NULL,
    comment_text text NOT NULL,
    stage_name character varying(30) NOT NULL,
    commented_by character varying(100) NOT NULL,
    commented_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT chk_comment_stage CHECK (((stage_name)::text = ANY (ARRAY[('INDEXING'::character varying)::text, ('DATA_ENTRY'::character varying)::text, ('QUALITY_CHECK'::character varying)::text, ('APPROVAL'::character varying)::text, ('FIRST_COLLECTION'::character varying)::text, ('INSTALLMENT_PAYMENT'::character varying)::text, ('GENERAL'::character varying)::text])))
);


--
-- Name: TABLE request_comments; Type: COMMENT; Schema: revival; Owner: -
--

COMMENT ON TABLE revival.request_comments IS 'Comments/notes added during revival workflow stages';


--
-- Name: request_stage_history; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.request_stage_history (
    history_id character varying NOT NULL,
    request_id character varying NOT NULL,
    stage_name character varying NOT NULL,
    stage_action character varying NOT NULL,
    comments text,
    action_at timestamp without time zone DEFAULT now() NOT NULL,
    action_by character varying NOT NULL,
    missing_documents text[]
);


--
-- Name: revival_calculations; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.revival_calculations (
    calculation_id character varying NOT NULL,
    request_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    date_of_revival date NOT NULL,
    number_of_installments integer NOT NULL,
    unpaid_premium_months integer NOT NULL,
    interest_rate numeric,
    monthly_premium numeric,
    premium_amount numeric NOT NULL,
    tax_on_premium numeric NOT NULL,
    total_renewal_amount numeric NOT NULL,
    installment_amount numeric NOT NULL,
    tax_on_unpaid_premium numeric NOT NULL,
    total_installment_amount numeric NOT NULL,
    grand_total_first_collection numeric NOT NULL,
    valid_until timestamp without time zone,
    calculated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: revival_request_workflow_state; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.revival_request_workflow_state (
    request_id character varying NOT NULL,
    workflow_id character varying NOT NULL,
    run_id character varying NOT NULL,
    current_status character varying NOT NULL,
    workflow_status character varying DEFAULT 'RUNNING'::character varying NOT NULL,
    sla_start_date timestamp without time zone,
    sla_end_date timestamp without time zone,
    sla_expired boolean DEFAULT false NOT NULL,
    first_collection_done boolean DEFAULT false NOT NULL,
    total_installments integer,
    installments_paid integer DEFAULT 0 NOT NULL,
    started_at timestamp without time zone DEFAULT now() NOT NULL,
    completed_at timestamp without time zone,
    last_updated timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: revival_requests; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.revival_requests (
    request_id character varying NOT NULL,
    ticket_id character varying NOT NULL,
    policy_number character varying NOT NULL,
    request_type character varying NOT NULL,
    current_status character varying NOT NULL,
    indexed_date timestamp without time zone,
    indexed_by character varying,
    data_entry_date timestamp without time zone,
    data_entry_by character varying,
    qc_complete_date timestamp without time zone,
    qc_by character varying,
    approval_date timestamp without time zone,
    approved_by character varying,
    completion_date timestamp without time zone,
    termination_date timestamp without time zone,
    withdrawal_date timestamp without time zone,
    number_of_installments integer NOT NULL,
    revival_amount numeric NOT NULL,
    installment_amount numeric NOT NULL,
    total_tax_on_unpaid numeric NOT NULL,
    first_collection_date timestamp without time zone,
    first_collection_done boolean DEFAULT false NOT NULL,
    blocking_new_collections boolean DEFAULT true NOT NULL,
    installments_paid integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    workflow_id character varying,
    run_id character varying,
    missing_documents_list text,
    documents text,
    request_owner character varying,
    previous_suspense_amount numeric(15,2) DEFAULT 0,
    suspense_adjusted boolean DEFAULT false,
    adjusted_revival_amount numeric(15,2) DEFAULT 0,
    sla_start_date timestamp without time zone,
    sla_end_date timestamp without time zone,
    sla_expired boolean DEFAULT false,
    interest numeric(15,2),
    sgst numeric(15,2),
    cgst numeric(15,2),
    rebate numeric(15,2),
    office_id character varying(50),
    user_id character varying(100),
    neft_details jsonb,
    revival_type character varying,
    qc_comments text,
    approval_comments text,
    medical_examiner_code character varying,
    medical_examiner_name character varying
);


--
-- Name: COLUMN revival_requests.revival_type; Type: COMMENT; Schema: revival; Owner: -
--

COMMENT ON COLUMN revival.revival_requests.revival_type IS 'Type of revival: installment or lumpsum';


--
-- Name: COLUMN revival_requests.qc_comments; Type: COMMENT; Schema: revival; Owner: -
--

COMMENT ON COLUMN revival.revival_requests.qc_comments IS 'Comments provided during quality check';


--
-- Name: COLUMN revival_requests.approval_comments; Type: COMMENT; Schema: revival; Owner: -
--

COMMENT ON COLUMN revival.revival_requests.approval_comments IS 'Comments provided during approval/rejection';


--
-- Name: revival_settings; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.revival_settings (
    setting_key character varying NOT NULL,
    setting_value character varying NOT NULL,
    is_configurable boolean DEFAULT false NOT NULL,
    description character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: revival_settings_audit; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.revival_settings_audit (
    audit_id character varying NOT NULL,
    setting_key character varying NOT NULL,
    old_value character varying NOT NULL,
    new_value character varying NOT NULL,
    changed_by character varying NOT NULL,
    changed_reason character varying,
    changed_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: status_change_history; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.status_change_history (
    history_id character varying NOT NULL,
    request_id character varying NOT NULL,
    from_status character varying,
    to_status character varying NOT NULL,
    changed_at timestamp without time zone DEFAULT now() NOT NULL,
    changed_by character varying NOT NULL,
    change_reason character varying
);


--
-- Name: workflow_executions; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.workflow_executions (
    execution_id character varying NOT NULL,
    request_id character varying NOT NULL,
    workflow_type character varying NOT NULL,
    workflow_id character varying NOT NULL,
    run_id character varying,
    started_at timestamp without time zone NOT NULL,
    completed_at timestamp without time zone,
    status character varying NOT NULL,
    parent_workflow_id character varying,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: workflow_signals_log; Type: TABLE; Schema: revival; Owner: -
--

CREATE TABLE revival.workflow_signals_log (
    log_id character varying NOT NULL,
    request_id character varying NOT NULL,
    workflow_id character varying NOT NULL,
    signal_name character varying NOT NULL,
    signal_payload jsonb,
    sent_at timestamp without time zone DEFAULT now() NOT NULL,
    sent_by character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: cheque_clearing_status cheque_clearing_status_pkey; Type: CONSTRAINT; Schema: collection; Owner: -
--

ALTER TABLE ONLY collection.cheque_clearing_status
    ADD CONSTRAINT cheque_clearing_status_pkey PRIMARY KEY (cheque_id);


--
-- Name: collection_batch_tracking collection_batch_tracking_pkey; Type: CONSTRAINT; Schema: collection; Owner: -
--

ALTER TABLE ONLY collection.collection_batch_tracking
    ADD CONSTRAINT collection_batch_tracking_pkey PRIMARY KEY (batch_id);


--
-- Name: payment_receipts payment_receipts_pkey; Type: CONSTRAINT; Schema: collection; Owner: -
--

ALTER TABLE ONLY collection.payment_receipts
    ADD CONSTRAINT payment_receipts_pkey PRIMARY KEY (receipt_id);


--
-- Name: payment_receipts payment_receipts_receipt_number_key; Type: CONSTRAINT; Schema: collection; Owner: -
--

ALTER TABLE ONLY collection.payment_receipts
    ADD CONSTRAINT payment_receipts_receipt_number_key UNIQUE (receipt_number);


--
-- Name: payment_transactions payment_transactions_pkey; Type: CONSTRAINT; Schema: collection; Owner: -
--

ALTER TABLE ONLY collection.payment_transactions
    ADD CONSTRAINT payment_transactions_pkey PRIMARY KEY (payment_id);


--
-- Name: suspense_accounts suspense_accounts_pkey; Type: CONSTRAINT; Schema: collection; Owner: -
--

ALTER TABLE ONLY collection.suspense_accounts
    ADD CONSTRAINT suspense_accounts_pkey PRIMARY KEY (suspense_id);


--
-- Name: configuration_audit configuration_audit_pkey; Type: CONSTRAINT; Schema: common; Owner: -
--

ALTER TABLE ONLY common.configuration_audit
    ADD CONSTRAINT configuration_audit_pkey PRIMARY KEY (audit_id);


--
-- Name: policies policies_pkey; Type: CONSTRAINT; Schema: common; Owner: -
--

ALTER TABLE ONLY common.policies
    ADD CONSTRAINT policies_pkey PRIMARY KEY (policy_number);


--
-- Name: system_configuration system_configuration_pkey; Type: CONSTRAINT; Schema: common; Owner: -
--

ALTER TABLE ONLY common.system_configuration
    ADD CONSTRAINT system_configuration_pkey PRIMARY KEY (config_key);


--
-- Name: tigerbeetle_accounts tigerbeetle_accounts_pkey; Type: CONSTRAINT; Schema: common; Owner: -
--

ALTER TABLE ONLY common.tigerbeetle_accounts
    ADD CONSTRAINT tigerbeetle_accounts_pkey PRIMARY KEY (account_id);


--
-- Name: tigerbeetle_accounts tigerbeetle_accounts_policy_number_key; Type: CONSTRAINT; Schema: common; Owner: -
--

ALTER TABLE ONLY common.tigerbeetle_accounts
    ADD CONSTRAINT tigerbeetle_accounts_policy_number_key UNIQUE (policy_number);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: document_references document_references_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.document_references
    ADD CONSTRAINT document_references_pkey PRIMARY KEY (reference_id);


--
-- Name: generated_letters generated_letters_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.generated_letters
    ADD CONSTRAINT generated_letters_pkey PRIMARY KEY (letter_id);


--
-- Name: installment_schedules installment_schedules_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.installment_schedules
    ADD CONSTRAINT installment_schedules_pkey PRIMARY KEY (schedule_id);


--
-- Name: request_comments request_comments_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.request_comments
    ADD CONSTRAINT request_comments_pkey PRIMARY KEY (comment_id);


--
-- Name: request_stage_history request_stage_history_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.request_stage_history
    ADD CONSTRAINT request_stage_history_pkey PRIMARY KEY (history_id);


--
-- Name: revival_calculations revival_calculations_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.revival_calculations
    ADD CONSTRAINT revival_calculations_pkey PRIMARY KEY (calculation_id);


--
-- Name: revival_request_workflow_state revival_request_workflow_state_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.revival_request_workflow_state
    ADD CONSTRAINT revival_request_workflow_state_pkey PRIMARY KEY (request_id);


--
-- Name: revival_requests revival_requests_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.revival_requests
    ADD CONSTRAINT revival_requests_pkey PRIMARY KEY (request_id);


--
-- Name: revival_requests revival_requests_ticket_id_key; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.revival_requests
    ADD CONSTRAINT revival_requests_ticket_id_key UNIQUE (ticket_id);


--
-- Name: revival_settings_audit revival_settings_audit_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.revival_settings_audit
    ADD CONSTRAINT revival_settings_audit_pkey PRIMARY KEY (audit_id);


--
-- Name: revival_settings revival_settings_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.revival_settings
    ADD CONSTRAINT revival_settings_pkey PRIMARY KEY (setting_key);


--
-- Name: status_change_history status_change_history_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.status_change_history
    ADD CONSTRAINT status_change_history_pkey PRIMARY KEY (history_id);


--
-- Name: workflow_executions workflow_executions_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.workflow_executions
    ADD CONSTRAINT workflow_executions_pkey PRIMARY KEY (execution_id);


--
-- Name: workflow_signals_log workflow_signals_log_pkey; Type: CONSTRAINT; Schema: revival; Owner: -
--

ALTER TABLE ONLY revival.workflow_signals_log
    ADD CONSTRAINT workflow_signals_log_pkey PRIMARY KEY (log_id);


--
-- Name: idx_audit_reversed_at; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_audit_reversed_at ON collection.suspense_reversal_audit USING btree (reversed_at);


--
-- Name: idx_audit_suspense; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_audit_suspense ON collection.suspense_reversal_audit USING btree (suspense_id);


--
-- Name: idx_batch_complete; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_batch_complete ON collection.collection_batch_tracking USING btree (collection_complete);


--
-- Name: idx_batch_request; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_batch_request ON collection.collection_batch_tracking USING btree (request_id);


--
-- Name: idx_cheque_payment; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_cheque_payment ON collection.cheque_clearing_status USING btree (payment_id);


--
-- Name: idx_cheque_policy; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_cheque_policy ON collection.cheque_clearing_status USING btree (policy_number);


--
-- Name: idx_cheque_status; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_cheque_status ON collection.cheque_clearing_status USING btree (clearance_status);


--
-- Name: idx_payment_batch; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_payment_batch ON collection.payment_transactions USING btree (collection_batch_id);


--
-- Name: idx_payment_cheque; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_payment_cheque ON collection.payment_transactions USING btree (cheque_id);


--
-- Name: idx_payment_collection_date; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_payment_collection_date ON collection.payment_transactions USING btree (collection_date);


--
-- Name: idx_payment_request; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_payment_request ON collection.payment_transactions USING btree (request_id);


--
-- Name: idx_payment_status; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_payment_status ON collection.payment_transactions USING btree (payment_status);


--
-- Name: idx_receipt_date; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_receipt_date ON collection.payment_receipts USING btree (receipt_date);


--
-- Name: idx_receipt_number; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_receipt_number ON collection.payment_receipts USING btree (receipt_number);


--
-- Name: idx_receipt_request; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_receipt_request ON collection.payment_receipts USING btree (request_id);


--
-- Name: idx_suspense_policy; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_suspense_policy ON collection.suspense_accounts USING btree (policy_number);


--
-- Name: idx_suspense_request; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_suspense_request ON collection.suspense_accounts USING btree (request_id);


--
-- Name: idx_suspense_reversed; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_suspense_reversed ON collection.suspense_accounts USING btree (is_reversed);


--
-- Name: idx_suspense_type; Type: INDEX; Schema: collection; Owner: -
--

CREATE INDEX idx_suspense_type ON collection.suspense_accounts USING btree (suspense_type);


--
-- Name: idx_audit_changed_at; Type: INDEX; Schema: common; Owner: -
--

CREATE INDEX idx_audit_changed_at ON common.configuration_audit USING btree (changed_at);


--
-- Name: idx_audit_config_key; Type: INDEX; Schema: common; Owner: -
--

CREATE INDEX idx_audit_config_key ON common.configuration_audit USING btree (config_key);


--
-- Name: idx_config_key; Type: INDEX; Schema: common; Owner: -
--

CREATE INDEX idx_config_key ON common.system_configuration USING btree (config_key);


--
-- Name: idx_policy_customer; Type: INDEX; Schema: common; Owner: -
--

CREATE INDEX idx_policy_customer ON common.policies USING btree (customer_id);


--
-- Name: idx_policy_status; Type: INDEX; Schema: common; Owner: -
--

CREATE INDEX idx_policy_status ON common.policies USING btree (policy_status);


--
-- Name: idx_tb_policy; Type: INDEX; Schema: common; Owner: -
--

CREATE UNIQUE INDEX idx_tb_policy ON common.tigerbeetle_accounts USING btree (policy_number);


--
-- Name: idx_audit_changed_at; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_audit_changed_at ON revival.revival_settings_audit USING btree (changed_at);


--
-- Name: idx_audit_key; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_audit_key ON revival.revival_settings_audit USING btree (setting_key);


--
-- Name: idx_comments_commented_at; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_comments_commented_at ON revival.request_comments USING btree (commented_at DESC);


--
-- Name: idx_comments_policy; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_comments_policy ON revival.request_comments USING btree (policy_number);


--
-- Name: idx_comments_request; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_comments_request ON revival.request_comments USING btree (request_id);


--
-- Name: idx_comments_stage; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_comments_stage ON revival.request_comments USING btree (stage_name);


--
-- Name: idx_comments_ticket; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_comments_ticket ON revival.request_comments USING btree (ticket_id);


--
-- Name: idx_execution_request; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_execution_request ON revival.workflow_executions USING btree (request_id);


--
-- Name: idx_execution_started_at; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_execution_started_at ON revival.workflow_executions USING btree (started_at);


--
-- Name: idx_execution_status; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_execution_status ON revival.workflow_executions USING btree (status);


--
-- Name: idx_execution_type; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_execution_type ON revival.workflow_executions USING btree (workflow_type);


--
-- Name: idx_install_schedule_due; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_install_schedule_due ON revival.installment_schedules USING btree (due_date);


--
-- Name: idx_install_schedule_policy; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_install_schedule_policy ON revival.installment_schedules USING btree (policy_number);


--
-- Name: idx_install_schedule_request; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_install_schedule_request ON revival.installment_schedules USING btree (request_id);


--
-- Name: idx_letter_request; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_letter_request ON revival.generated_letters USING btree (request_id);


--
-- Name: idx_letter_type; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_letter_type ON revival.generated_letters USING btree (letter_type);


--
-- Name: idx_ref_request; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_ref_request ON revival.document_references USING btree (request_id);


--
-- Name: idx_ref_type; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_ref_type ON revival.document_references USING btree (document_type);


--
-- Name: idx_revival_calc_policy; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_revival_calc_policy ON revival.revival_calculations USING btree (policy_number);


--
-- Name: idx_revival_calc_request_id; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_revival_calc_request_id ON revival.revival_calculations USING btree (request_id);


--
-- Name: idx_revival_current_status; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_revival_current_status ON revival.revival_requests USING btree (current_status);


--
-- Name: idx_revival_policy_number; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_revival_policy_number ON revival.revival_requests USING btree (policy_number);


--
-- Name: idx_revival_ticket_id; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_revival_ticket_id ON revival.revival_requests USING btree (ticket_id);


--
-- Name: idx_revival_workflow_id; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_revival_workflow_id ON revival.revival_requests USING btree (workflow_id);


--
-- Name: idx_settings_key; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_settings_key ON revival.revival_settings USING btree (setting_key);


--
-- Name: idx_signal_name; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_signal_name ON revival.workflow_signals_log USING btree (signal_name);


--
-- Name: idx_signal_request; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_signal_request ON revival.workflow_signals_log USING btree (request_id);


--
-- Name: idx_signal_sent_at; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_signal_sent_at ON revival.workflow_signals_log USING btree (sent_at);


--
-- Name: idx_stage_history_action_at; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_stage_history_action_at ON revival.request_stage_history USING btree (action_at);


--
-- Name: idx_stage_history_request; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_stage_history_request ON revival.request_stage_history USING btree (request_id);


--
-- Name: idx_stage_history_stage; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_stage_history_stage ON revival.request_stage_history USING btree (stage_name);


--
-- Name: idx_status_history_changed_at; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_status_history_changed_at ON revival.status_change_history USING btree (changed_at);


--
-- Name: idx_status_history_request; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_status_history_request ON revival.status_change_history USING btree (request_id);


--
-- Name: idx_workflow_state_status; Type: INDEX; Schema: revival; Owner: -
--

CREATE INDEX idx_workflow_state_status ON revival.revival_request_workflow_state USING btree (current_status);


--
-- PostgreSQL database dump complete
--

\unrestrict dN5DqlqSrhOZrJoZVewEougEt67A3lnfoO7Gxb31ZKvODfn9SHEMHuUjQsbpzbQ


ALTER TABLE policy_status_history 
ALTER COLUMN request_id TYPE VARCHAR(36);


ALTER TABLE policy_mgmt.policy
ADD COLUMN mobile_number VARCHAR(15);



ALTER TABLE policy_mgmt.policy
ADD COLUMN email varchar(255);


ALTER TABLE policy_mgmt.policy
ADD COLUMN address_type varchar(255);


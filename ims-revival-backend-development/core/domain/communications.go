package domain

import "github.com/volatiletech/null/v9"

// EmpCommunicationDetails represents a record from the employee_communication_details table
type EmpCommunicationDetails struct {
	CommunicationID     null.Int32  `json:"communication_id" db:"communication_id"`
	EmployeeID          null.Int64  `json:"employee_id" db:"employee_id"`
	CommunicationAddr1  null.String `json:"communication_address_1" db:"communication_address_1"`
	CommunicationAddr2  null.String `json:"communication_address_2" db:"communication_address_2"`
	CommunicationAddr3  null.String `json:"communication_address_3" db:"communication_address_3"`
	CommunicationPIN    null.Int32  `json:"communication_pin" db:"communication_pin"`
	IndiaPostEmailID    null.String `json:"india_post_email_id" db:"india_post_email_id"`
	PersonalEmailID     null.String `json:"personal_email_id" db:"personal_email_id"`
	MobileNo            null.Int64  `json:"mobile_no" db:"mobile_no"`
	AadhaarRefNumber    null.String `json:"aadhaar_ref_number" db:"aadhaar_ref_number"`
	PANNumber           null.String `json:"pan_number" db:"pan_number"`
	ApproverPostID      null.String `json:"approver_post_id" db:"approver_post_id"`
	Status              null.String `json:"status" db:"status"`
	CreatedBy           null.String `json:"created_by" db:"created_by"`
	CreatedDate         null.Time   `json:"created_date" db:"created_date"`
	UpdatedBy           null.String `json:"updated_by" db:"updated_by"`
	UpdatedDate         null.Time   `json:"updated_date" db:"updated_date"`
	ApprovedBy          null.String `json:"approved_by" db:"approved_by"`
	ApprovedDate        null.Time   `json:"approved_date" db:"approved_date"`
	Remarks             null.String `json:"remarks" db:"remarks"`
	UserRemarks         null.String `json:"user_remarks" db:"user_remarks"`
	FwdAuthRemarks      null.String `json:"fwd_auth_remarks" db:"fwd_auth_remarks"`
	ApprRemarks         null.String `json:"approver_remarks" db:"approve_auth_remarks"`
	ForwardPostID       null.String `json:"forward_post_id" db:"forward_post_id"`
	EmployeeName        null.String `json:"employee_name" db:"employee_name"`
	AdminOffice         null.Int64  `json:"admin_office" db:"admin_office"`
	OfficeOfWorking     null.String `json:"office_of_working" db:"office_of_working" `
	EmployeeDesignation null.String `json:"employee_designation" db:"employee_designation" `
	ApproverOfficeID    null.Int64  `json:"approver_office_id" db:"approver_office_id"`
	EmpPostID           null.Int64  `json:"emp_post_id" db:"emp_post_id"`
	OfficeID            null.Int64  `json:"office_id" db:"office_id"`
	OfficeTypeCode      null.String `json:"office_type_code" db:"office_type_code"`
	ResourceID          null.Int64  `json:"resource_id" db:"resource_id"`
	ResourceStatus      null.String `json:"resource_status" db:"resource_status"`
}

type CombinedStructForLMS struct {
	CommResp EmpCommunicationDetails `json:"communications"`
	EmpResp  FetchEmployee           `json:"emp_details"`
}

type MobileStatusReport struct {
	CircleOfficeID      null.Int     `json:"circle_office_id" db:"circle_office_id"`
	CircleName          null.String  `json:"circle_name" db:"circle_name"`
	TotalEmpCount       null.Int     `json:"total_emp_count" db:"total_emp_count"`
	EmpWithValidMobile  null.Int     `json:"emp_with_valid_mobile" db:"emp_with_valid_mobile"`
	YetToUpdate         null.Int     `json:"yet_to_update" db:"yet_to_update"`
	PercentageCompleted null.Float64 `json:"percentage_completed" db:"percentage_completed"`
}

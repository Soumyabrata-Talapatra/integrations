package domain

import (
	"github.com/volatiletech/null/v9"
)

type CommunityEnum string

const (
	UR  CommunityEnum = "UR"
	OBC CommunityEnum = "OBC"
	SC  CommunityEnum = "SC"
	ST  CommunityEnum = "ST"
	EWS CommunityEnum = "EWS"
	CNA CommunityEnum = "NA"
)

type GenderEnum string

const (
	Male        GenderEnum = "Male"
	Female      GenderEnum = "Female"
	TransMale   GenderEnum = "Transgender-Male"
	TransFemale GenderEnum = "Transgender-Female"
	GNA         GenderEnum = "NA"
)

type GroupEnum string

const (
	GroupA       GroupEnum = "Group A"
	GroupBGaz    GroupEnum = "Group B Gazetted"
	GroupBNonGaz GroupEnum = "Group B Non Gazetted"
	GroupC       GroupEnum = "Group C"
	GroupGDS     GroupEnum = "GDS"
)

type MaritalEnum string

const (
	Married   MaritalEnum = "Married"
	UnMarried MaritalEnum = "UnMarried"
	Divorcee  MaritalEnum = "Divorcee"
	Widower   MaritalEnum = "Widower"
	Widow     MaritalEnum = "Widow"
	MNA       MaritalEnum = "NA"
)

type EmpStatusEnum string

const (
	Active         EmpStatusEnum = "Active"
	Suspension     EmpStatusEnum = "Suspension"
	Pending        EmpStatusEnum = "Pending"
	Inactive       EmpStatusEnum = "Inactive"
	Secondment     EmpStatusEnum = "Secondment"
	RP             EmpStatusEnum = "RP"
	Death          EmpStatusEnum = "Death"
	VRS            EmpStatusEnum = "VRS"
	IVP            EmpStatusEnum = "IVP"
	CR             EmpStatusEnum = "CR"
	CRP            EmpStatusEnum = "CRP"
	CP             EmpStatusEnum = "CP"
	CPP            EmpStatusEnum = "CPP"
	Superannuation EmpStatusEnum = "Superannuation"
	Resignation    EmpStatusEnum = "Resignation"
)

type EmpTypeEnum string

const (
	DOP EmpTypeEnum = "DOP"
	GDS EmpTypeEnum = "GDS"
)

type RecruitEnum string

const (
	Sports        RecruitEnum = "Sports"
	DR            RecruitEnum = "DR"
	DP            RecruitEnum = "DP"
	Compassionate RecruitEnum = "Compassionate"
)

type TaxEnum string

const (
	Old TaxEnum = "Old"
	New TaxEnum = "New"
)

type DepTypeEnum string

const (
	APS             DepTypeEnum = "APS"
	POST            DepTypeEnum = "DOP"
	IPPB            DepTypeEnum = "IPPB"
	OtherDepartment DepTypeEnum = "OtherDepartment"
)

type PensionSchemeEnum string

const (
	GPF     PensionSchemeEnum = "GPF"
	NPS     PensionSchemeEnum = "NPS"
	UPS     PensionSchemeEnum = "UPS"
	SDBS    PensionSchemeEnum = "SDBS"
	NonSDBS PensionSchemeEnum = "Non-SDBS"
	NA      PensionSchemeEnum = "NA"
)

// EmpAwardDetails represents the details of an award
type EmpAwardDetails struct {
	AwardDetailsID   null.Int     `json:"award_details_id" db:"award_details_id"`
	EmployeeID       null.Int64   `json:"employee_id" db:"employee_id"`
	AwardName        null.String  `json:"award_name" db:"award_name"`
	AwardType        null.String  `json:"award_type" db:"award_type"`
	AwardCategory    null.String  `json:"award_category" db:"award_category"`
	AwardDescription null.String  `json:"award_description" db:"award_description"`
	CertificateNo    null.String  `json:"certificate_no" db:"certificate_no"`
	MonetaryBenefit  null.Float64 `json:"monetary_benefit" db:"monetary_benefit"`
	AwardIssueDate   null.Time    `json:"award_issue_date" db:"award_issue_date"`
	AwardForYear     null.Int     `json:"award_for_year" db:"award_for_year"`
	ApproverPostID   null.String  `json:"approver_post_id" db:"approver_post_id"`
	Status           null.String  `json:"status" db:"status"`
	CreatedBy        null.String  `json:"created_by" db:"created_by"`
	CreatedDate      null.Time    `json:"created_date" db:"created_date"`
	UpdatedBy        null.String  `json:"updated_by" db:"updated_by"`
	UpdatedDate      null.Time    `json:"updated_date" db:"updated_date"`
	ApprovedBy       null.String  `json:"approved_by" db:"approved_by"`
	ApprovedDate     null.Time    `json:"approved_date" db:"approved_date"`
	Remarks          null.String  `json:"remarks" db:"remarks"`
	UserRemarks      null.String  `json:"user_remarks" db:"user_remarks"`
	FwdAuthRemarks   null.String  `json:"fwd_auth_remarks" db:"fwd_auth_remarks"`
	ApprRemarks      null.String  `json:"approver_remarks" db:"approve_auth_remarks"`
	ForwardPostID    null.String  `json:"forward_post_id" db:"forward_post_id"`
	EmployeeName     null.String  `json:"employee_name" db:"employee_name"`
	AdminOffice      null.Int64   `json:"admin_office" db:"admin_office"`
	AwardID          null.Int32   `json:"award_id" validate:"required"`
	OffOfWorking     null.String  `json:"office_of_working" db:"office_of_working"`
	EmpDesignation   null.String  `json:"employee_designation" db:"employee_designation"`
	FileName         string       `json:"file_name"`
	ApproverOfficeID null.Int64   `json:"approver_office_id" db:"approver_office_id"`
	EmpPostID        null.Int64   `json:"emp_post_id" db:"emp_post_id"`
	OfficeID         null.Int64   `json:"office_id" db:"office_id"`
	OfficeTypeCode   null.String  `json:"office_type_code" db:"office_type_code"`
}

type AwardsCreateResponse struct {
	AwardDetailsID null.Int    `json:"award_details_id" db:"award_details_id"`
	EmployeeID     null.Int64  `json:"employee_id" db:"employee_id"`
	AwardName      null.String `json:"award_name" db:"award_name"`
	AwardType      null.String `json:"award_type" db:"award_type"`
	UserRemarks    null.String `json:"user_remarks" db:"user_remarks"`
	Status         null.String `json:"status" db:"status"`
}

type FetchEmployee struct {
	EmployeeID               null.Int64   `json:"employee_id" db:"employee_id"`
	FirstName                null.String  `json:"employee_first_name" db:"employee_first_name"`
	GroupPost                GroupEnum    `json:"group_post" db:"group_post"`
	Cadre                    null.String  `json:"cadre" db:"cadre"`
	EmployeeDesignation      null.String  `json:"employee_designation" db:"employee_designation"`
	OfficeOfWorking          null.String  `json:"office_of_working" db:"office_of_working"`
	PostID                   null.Int64   `json:"post_id" db:"post_id"`
	OfficeID                 null.Int64   `json:"office_id" db:"office_id"`
	EmploymentStatus         null.String  `json:"employment_status" db:"employment_status"`
	EmployeeType             EmpTypeEnum  `json:"employee_type" db:"employee_type"`
	Gender                   null.String  `json:"gender" db:"gender"`
	CircleOfficeID           null.Int64   `json:"circle_office_id" db:"circle_office_id"`
	DateOfJoinInDepartment   null.Time    `json:"date_of_join_in_department" db:"date_of_join_in_department"`
	DateOfJoinInPresentCadre null.Time    `json:"date_of_join_in_present_cadre" db:"date_of_join_in_present_cadre"`
	MaritalStatus            MaritalEnum  `json:"marital_status" db:"marital_status"`
	RecruitmentMode          RecruitEnum  `json:"recruitment_mode" db:"recruitment_mode"`
	OfficeType               null.String  `json:"office_type" db:"office_type"`
	ReportingOfficeID        null.Int64   `json:"reporting_office_id" db:"reporting_office_id"`
	ReportingAuthorityPostID null.Int64   `json:"reporting_authority_post_id" db:"reporting_authority_post_id"`
	CadreID                  null.Int     `json:"cadre_id" db:"cadre_id"`
	DDOID                    null.Int64   `json:"ddo_office_id" db:"ddo_office_id"`
	SubDivID                 null.Int64   `json:"sub_division_office_id" db:"sub_division_office_id"`
	DivID                    null.Int64   `json:"division_office_id" db:"division_office_id"`
	RegionID                 null.Int64   `json:"region_office_id"  db:"region_office_id" `
	PayScaleLevel            null.String  `json:"pay_scale_level" db:"pay_scale_level"`
	PayScaleIndex            null.Int32   `json:"pay_scale_index" db:"pay_scale_index"`
	BasicPay                 null.Float32 `json:"basic_pay" db:"basic_pay"`
	DivName                  null.String  `json:"division_name" db:"division_name"`
	RegName                  null.String  `json:"region_name" db:"region_name"`
	CircleName               null.String  `json:"circle_name" db:"circle_name"`
	GroupID                  null.Int32   `json:"group_id" db:"group_id"`
	PWD                      null.Bool    `json:"pwd" db:"pwd"`
	PWDSubCat                null.String  `json:"pwd_sub_cat" db:"pwd_sub_cat"`
	PWDPercentage            null.Float32 `json:"pwd_percentage" db:"pwd_percentage"`
}

type EmpDetailsResponse struct {
	DetailsID  uint64 `json:"details_id" db:"details_id"`
	EmployeeID int64  `json:"employee_id" db:"employee_id"`
	Status     string `json:"status" db:"status"`
	Remarks    string `json:"remarks" db:"remarks"`
}

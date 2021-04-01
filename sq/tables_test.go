package sq

type APPLICATIONS struct {
	TableInfo
	APPLICATION_DATA     JSONField
	APPLICATION_FORM_ID  NumberField
	APPLICATION_ID       NumberField
	COHORT               StringField
	CREATED_AT           TimeField
	CREATOR_USER_ROLE_ID NumberField
	DELETED_AT           TimeField
	MAGICSTRING          StringField
	PROJECT_IDEA         StringField
	PROJECT_LEVEL        StringField
	STATUS               StringField
	SUBMITTED            BooleanField
	TEAM_ID              NumberField
	TEAM_NAME            StringField
	UPDATED_AT           TimeField
}

func NEW_APPLICATIONS(alias string) APPLICATIONS {
	tbl := APPLICATIONS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "applications",
		Alias:  alias,
	}}
	tbl.APPLICATION_DATA = NewJSONField("application_data", tbl.TableInfo)
	tbl.APPLICATION_FORM_ID = NewNumberField("application_form_id", tbl.TableInfo)
	tbl.APPLICATION_ID = NewNumberField("application_id", tbl.TableInfo)
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.CREATOR_USER_ROLE_ID = NewNumberField("creator_user_role_id", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.MAGICSTRING = NewStringField("magicstring", tbl.TableInfo)
	tbl.PROJECT_IDEA = NewStringField("project_idea", tbl.TableInfo)
	tbl.PROJECT_LEVEL = NewStringField("project_level", tbl.TableInfo)
	tbl.STATUS = NewStringField("status", tbl.TableInfo)
	tbl.SUBMITTED = NewBooleanField("submitted", tbl.TableInfo)
	tbl.TEAM_ID = NewNumberField("team_id", tbl.TableInfo)
	tbl.TEAM_NAME = NewStringField("team_name", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type APPLICATIONS_STATUS_ENUM struct {
	TableInfo
	STATUS StringField
}

func NEW_APPLICATIONS_STATUS_ENUM(alias string) APPLICATIONS_STATUS_ENUM {
	tbl := APPLICATIONS_STATUS_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "applications_status_enum",
		Alias:  alias,
	}}
	tbl.STATUS = NewStringField("status", tbl.TableInfo)
	return tbl
}

type COHORT_ENUM struct {
	TableInfo
	COHORT          StringField
	INSERTION_ORDER NumberField
}

func NEW_COHORT_ENUM(alias string) COHORT_ENUM {
	tbl := COHORT_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "cohort_enum",
		Alias:  alias,
	}}
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.INSERTION_ORDER = NewNumberField("insertion_order", tbl.TableInfo)
	return tbl
}

type TABLE_FEEDBACK_ON_TEAMS struct {
	TableInfo
	CREATED_AT          TimeField
	DELETED_AT          TimeField
	EVALUATEE_TEAM_ID   NumberField
	EVALUATOR_TEAM_ID   NumberField
	FEEDBACK_DATA       JSONField
	FEEDBACK_FORM_ID    NumberField
	FEEDBACK_ID_ON_TEAM NumberField
	OVERRIDE_OPEN       BooleanField
	SUBMITTED           BooleanField
	UPDATED_AT          TimeField
}

func FEEDBACK_ON_TEAMS() TABLE_FEEDBACK_ON_TEAMS {
	tbl := TABLE_FEEDBACK_ON_TEAMS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "feedback_on_teams",
	}}
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.EVALUATEE_TEAM_ID = NewNumberField("evaluatee_team_id", tbl.TableInfo)
	tbl.EVALUATOR_TEAM_ID = NewNumberField("evaluator_team_id", tbl.TableInfo)
	tbl.FEEDBACK_DATA = NewJSONField("feedback_data", tbl.TableInfo)
	tbl.FEEDBACK_FORM_ID = NewNumberField("feedback_form_id", tbl.TableInfo)
	tbl.FEEDBACK_ID_ON_TEAM = NewNumberField("feedback_id_on_team", tbl.TableInfo)
	tbl.OVERRIDE_OPEN = NewBooleanField("override_open", tbl.TableInfo)
	tbl.SUBMITTED = NewBooleanField("submitted", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type TABLE_FEEDBACK_ON_USERS struct {
	TableInfo
	CREATED_AT             TimeField
	DELETED_AT             TimeField
	EVALUATEE_USER_ROLE_ID NumberField
	EVALUATOR_TEAM_ID      NumberField
	FEEDBACK_DATA          JSONField
	FEEDBACK_FORM_ID       NumberField
	FEEDBACK_ID_ON_USER    NumberField
	OVERRIDE_OPEN          BooleanField
	SUBMITTED              BooleanField
	UPDATED_AT             TimeField
}

func FEEDBACK_ON_USERS() TABLE_FEEDBACK_ON_USERS {
	tbl := TABLE_FEEDBACK_ON_USERS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "feedback_on_users",
	}}
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.EVALUATEE_USER_ROLE_ID = NewNumberField("evaluatee_user_role_id", tbl.TableInfo)
	tbl.EVALUATOR_TEAM_ID = NewNumberField("evaluator_team_id", tbl.TableInfo)
	tbl.FEEDBACK_DATA = NewJSONField("feedback_data", tbl.TableInfo)
	tbl.FEEDBACK_FORM_ID = NewNumberField("feedback_form_id", tbl.TableInfo)
	tbl.FEEDBACK_ID_ON_USER = NewNumberField("feedback_id_on_user", tbl.TableInfo)
	tbl.OVERRIDE_OPEN = NewBooleanField("override_open", tbl.TableInfo)
	tbl.SUBMITTED = NewBooleanField("submitted", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type TABLE_FORMS struct {
	TableInfo
	CREATED_AT TimeField
	DELETED_AT TimeField
	FORM_ID    NumberField
	NAME       StringField
	PERIOD_ID  NumberField
	QUESTIONS  JSONField
	SUBSECTION StringField
	UPDATED_AT TimeField
}

func FORMS() TABLE_FORMS {
	tbl := TABLE_FORMS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "forms",
	}}
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.FORM_ID = NewNumberField("form_id", tbl.TableInfo)
	tbl.NAME = NewStringField("name", tbl.TableInfo)
	tbl.PERIOD_ID = NewNumberField("period_id", tbl.TableInfo)
	tbl.QUESTIONS = NewJSONField("questions", tbl.TableInfo)
	tbl.SUBSECTION = NewStringField("subsection", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type TABLE_FORMS_AUTHORIZED_ROLES struct {
	TableInfo
	FORM_ID NumberField
	ROLE    StringField
}

func FORMS_AUTHORIZED_ROLES() TABLE_FORMS_AUTHORIZED_ROLES {
	tbl := TABLE_FORMS_AUTHORIZED_ROLES{TableInfo: TableInfo{
		Schema: "public",
		Name:   "forms_authorized_roles",
	}}
	tbl.FORM_ID = NewNumberField("form_id", tbl.TableInfo)
	tbl.ROLE = NewStringField("role", tbl.TableInfo)
	return tbl
}

type TABLE_MEDIA struct {
	TableInfo
	CREATED_AT  TimeField
	DATA        BlobField
	DELETED_AT  TimeField
	DESCRIPTION StringField
	NAME        StringField
	TYPE        StringField
	UPDATED_AT  TimeField
}

func MEDIA() TABLE_MEDIA {
	tbl := TABLE_MEDIA{TableInfo: TableInfo{
		Schema: "public",
		Name:   "media",
	}}
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DATA = NewBlobField("data", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.DESCRIPTION = NewStringField("description", tbl.TableInfo)
	tbl.NAME = NewStringField("name", tbl.TableInfo)
	tbl.TYPE = NewStringField("type", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type TABLE_MILESTONE_ENUM struct {
	TableInfo
	MILESTONE StringField
}

func MILESTONE_ENUM() TABLE_MILESTONE_ENUM {
	tbl := TABLE_MILESTONE_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "milestone_enum",
	}}
	tbl.MILESTONE = NewStringField("milestone", tbl.TableInfo)
	return tbl
}

type TABLE_MIME_TYPE_ENUM struct {
	TableInfo
	TYPE StringField
}

func MIME_TYPE_ENUM() TABLE_MIME_TYPE_ENUM {
	tbl := TABLE_MIME_TYPE_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "mime_type_enum",
	}}
	tbl.TYPE = NewStringField("type", tbl.TableInfo)
	return tbl
}

type TABLE_PERIODS struct {
	TableInfo
	COHORT     StringField
	CREATED_AT TimeField
	DELETED_AT TimeField
	END_AT     TimeField
	MILESTONE  StringField
	PERIOD_ID  NumberField
	STAGE      StringField
	START_AT   TimeField
	UPDATED_AT TimeField
}

func PERIODS() TABLE_PERIODS {
	tbl := TABLE_PERIODS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "periods",
	}}
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.END_AT = NewTimeField("end_at", tbl.TableInfo)
	tbl.MILESTONE = NewStringField("milestone", tbl.TableInfo)
	tbl.PERIOD_ID = NewNumberField("period_id", tbl.TableInfo)
	tbl.STAGE = NewStringField("stage", tbl.TableInfo)
	tbl.START_AT = NewTimeField("start_at", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type TABLE_PROJECT_CATEGORY_ENUM struct {
	TableInfo
	PROJECT_CATEGORY StringField
}

func PROJECT_CATEGORY_ENUM() TABLE_PROJECT_CATEGORY_ENUM {
	tbl := TABLE_PROJECT_CATEGORY_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "project_category_enum",
	}}
	tbl.PROJECT_CATEGORY = NewStringField("project_category", tbl.TableInfo)
	return tbl
}

type TABLE_PROJECT_LEVEL_ENUM struct {
	TableInfo
	PROJECT_LEVEL StringField
}

func PROJECT_LEVEL_ENUM() TABLE_PROJECT_LEVEL_ENUM {
	tbl := TABLE_PROJECT_LEVEL_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "project_level_enum",
	}}
	tbl.PROJECT_LEVEL = NewStringField("project_level", tbl.TableInfo)
	return tbl
}

type TABLE_ROLE_ENUM struct {
	TableInfo
	ROLE StringField
}

func ROLE_ENUM() TABLE_ROLE_ENUM {
	tbl := TABLE_ROLE_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "role_enum",
	}}
	tbl.ROLE = NewStringField("role", tbl.TableInfo)
	return tbl
}

type TABLE_SESSIONS struct {
	TableInfo
	CREATED_AT TimeField
	HASH       StringField
	USER_ID    NumberField
}

func SESSIONS() TABLE_SESSIONS {
	tbl := TABLE_SESSIONS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "sessions",
	}}
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.HASH = NewStringField("hash", tbl.TableInfo)
	tbl.USER_ID = NewNumberField("user_id", tbl.TableInfo)
	return tbl
}

type TABLE_STAGE_ENUM struct {
	TableInfo
	STAGE StringField
}

func STAGE_ENUM() TABLE_STAGE_ENUM {
	tbl := TABLE_STAGE_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "stage_enum",
	}}
	tbl.STAGE = NewStringField("stage", tbl.TableInfo)
	return tbl
}

type TABLE_SUBMISSIONS struct {
	TableInfo
	CREATED_AT         TimeField
	DELETED_AT         TimeField
	OVERRIDE_OPEN      BooleanField
	POSTER             StringField
	README             StringField
	SUBMISSION_DATA    JSONField
	SUBMISSION_FORM_ID NumberField
	SUBMISSION_ID      NumberField
	SUBMITTED          BooleanField
	TEAM_ID            NumberField
	UPDATED_AT         TimeField
	VIDEO              StringField
}

func SUBMISSIONS() TABLE_SUBMISSIONS {
	tbl := TABLE_SUBMISSIONS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "submissions",
	}}
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.OVERRIDE_OPEN = NewBooleanField("override_open", tbl.TableInfo)
	tbl.POSTER = NewStringField("poster", tbl.TableInfo)
	tbl.README = NewStringField("readme", tbl.TableInfo)
	tbl.SUBMISSION_DATA = NewJSONField("submission_data", tbl.TableInfo)
	tbl.SUBMISSION_FORM_ID = NewNumberField("submission_form_id", tbl.TableInfo)
	tbl.SUBMISSION_ID = NewNumberField("submission_id", tbl.TableInfo)
	tbl.SUBMITTED = NewBooleanField("submitted", tbl.TableInfo)
	tbl.TEAM_ID = NewNumberField("team_id", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	tbl.VIDEO = NewStringField("video", tbl.TableInfo)
	return tbl
}

type TABLE_SUBMISSIONS_CATEGORIES struct {
	TableInfo
	CATEGORY      StringField
	SUBMISSION_ID NumberField
}

func SUBMISSIONS_CATEGORIES() TABLE_SUBMISSIONS_CATEGORIES {
	tbl := TABLE_SUBMISSIONS_CATEGORIES{TableInfo: TableInfo{
		Schema: "public",
		Name:   "submissions_categories",
	}}
	tbl.CATEGORY = NewStringField("category", tbl.TableInfo)
	tbl.SUBMISSION_ID = NewNumberField("submission_id", tbl.TableInfo)
	return tbl
}

type TABLE_TEAM_EVALUATION_PAIRS struct {
	TableInfo
	EVALUATEE_TEAM_ID NumberField
	EVALUATOR_TEAM_ID NumberField
}

func TEAM_EVALUATION_PAIRS() TABLE_TEAM_EVALUATION_PAIRS {
	tbl := TABLE_TEAM_EVALUATION_PAIRS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "team_evaluation_pairs",
	}}
	tbl.EVALUATEE_TEAM_ID = NewNumberField("evaluatee_team_id", tbl.TableInfo)
	tbl.EVALUATOR_TEAM_ID = NewNumberField("evaluator_team_id", tbl.TableInfo)
	return tbl
}

type TABLE_TEAM_EVALUATIONS struct {
	TableInfo
	CREATED_AT              TimeField
	DELETED_AT              TimeField
	EVALUATEE_SUBMISSION_ID NumberField
	EVALUATION_DATA         JSONField
	EVALUATION_FORM_ID      NumberField
	EVALUATOR_TEAM_ID       NumberField
	OVERRIDE_OPEN           BooleanField
	SUBMITTED               BooleanField
	TEAM_EVALUATION_ID      NumberField
	UPDATED_AT              TimeField
}

func TEAM_EVALUATIONS() TABLE_TEAM_EVALUATIONS {
	tbl := TABLE_TEAM_EVALUATIONS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "team_evaluations",
	}}
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.EVALUATEE_SUBMISSION_ID = NewNumberField("evaluatee_submission_id", tbl.TableInfo)
	tbl.EVALUATION_DATA = NewJSONField("evaluation_data", tbl.TableInfo)
	tbl.EVALUATION_FORM_ID = NewNumberField("evaluation_form_id", tbl.TableInfo)
	tbl.EVALUATOR_TEAM_ID = NewNumberField("evaluator_team_id", tbl.TableInfo)
	tbl.OVERRIDE_OPEN = NewBooleanField("override_open", tbl.TableInfo)
	tbl.SUBMITTED = NewBooleanField("submitted", tbl.TableInfo)
	tbl.TEAM_EVALUATION_ID = NewNumberField("team_evaluation_id", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type TABLE_TEAMS struct {
	TableInfo
	ADVISER_USER_ROLE_ID NumberField
	COHORT               StringField
	CREATED_AT           TimeField
	DELETED_AT           TimeField
	MENTOR_USER_ROLE_ID  NumberField
	PROJECT_IDEA         StringField
	PROJECT_LEVEL        StringField
	STATUS               StringField
	TEAM_DATA            JSONField
	TEAM_ID              NumberField
	TEAM_NAME            StringField
	UPDATED_AT           TimeField
}

func TEAMS() TABLE_TEAMS {
	tbl := TABLE_TEAMS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "teams",
	}}
	tbl.ADVISER_USER_ROLE_ID = NewNumberField("adviser_user_role_id", tbl.TableInfo)
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.MENTOR_USER_ROLE_ID = NewNumberField("mentor_user_role_id", tbl.TableInfo)
	tbl.PROJECT_IDEA = NewStringField("project_idea", tbl.TableInfo)
	tbl.PROJECT_LEVEL = NewStringField("project_level", tbl.TableInfo)
	tbl.STATUS = NewStringField("status", tbl.TableInfo)
	tbl.TEAM_DATA = NewJSONField("team_data", tbl.TableInfo)
	tbl.TEAM_ID = NewNumberField("team_id", tbl.TableInfo)
	tbl.TEAM_NAME = NewStringField("team_name", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type TABLE_TEAMS_STATUS_ENUM struct {
	TableInfo
	STATUS StringField
}

func TEAMS_STATUS_ENUM() TABLE_TEAMS_STATUS_ENUM {
	tbl := TABLE_TEAMS_STATUS_ENUM{TableInfo: TableInfo{
		Schema: "public",
		Name:   "teams_status_enum",
	}}
	tbl.STATUS = NewStringField("status", tbl.TableInfo)
	return tbl
}

type TABLE_USER_EVALUATIONS struct {
	TableInfo
	CREATED_AT              TimeField
	DELETED_AT              TimeField
	EVALUATEE_SUBMISSION_ID NumberField
	EVALUATION_DATA         JSONField
	EVALUATION_FORM_ID      NumberField
	EVALUATOR_USER_ROLE_ID  NumberField
	OVERRIDE_OPEN           BooleanField
	SUBMITTED               BooleanField
	UPDATED_AT              TimeField
	USER_EVALUATION_ID      NumberField
}

func USER_EVALUATIONS() TABLE_USER_EVALUATIONS {
	tbl := TABLE_USER_EVALUATIONS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "user_evaluations",
	}}
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.EVALUATEE_SUBMISSION_ID = NewNumberField("evaluatee_submission_id", tbl.TableInfo)
	tbl.EVALUATION_DATA = NewJSONField("evaluation_data", tbl.TableInfo)
	tbl.EVALUATION_FORM_ID = NewNumberField("evaluation_form_id", tbl.TableInfo)
	tbl.EVALUATOR_USER_ROLE_ID = NewNumberField("evaluator_user_role_id", tbl.TableInfo)
	tbl.OVERRIDE_OPEN = NewBooleanField("override_open", tbl.TableInfo)
	tbl.SUBMITTED = NewBooleanField("submitted", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	tbl.USER_EVALUATION_ID = NewNumberField("user_evaluation_id", tbl.TableInfo)
	return tbl
}

type TABLE_USER_ROLES struct {
	TableInfo
	COHORT       StringField
	CREATED_AT   TimeField
	DELETED_AT   TimeField
	ROLE         StringField
	UPDATED_AT   TimeField
	USER_ID      NumberField
	USER_ROLE_ID NumberField
}

func USER_ROLES() TABLE_USER_ROLES {
	tbl := TABLE_USER_ROLES{TableInfo: TableInfo{
		Schema: "public",
		Name:   "user_roles",
	}}
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.ROLE = NewStringField("role", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	tbl.USER_ID = NewNumberField("user_id", tbl.TableInfo)
	tbl.USER_ROLE_ID = NewNumberField("user_role_id", tbl.TableInfo)
	return tbl
}

type TABLE_USER_ROLES_APPLICANTS struct {
	TableInfo
	APPLICANT_DATA    JSONField
	APPLICANT_FORM_ID NumberField
	APPLICATION_ID    NumberField
	USER_ROLE_ID      NumberField
}

func USER_ROLES_APPLICANTS() TABLE_USER_ROLES_APPLICANTS {
	tbl := TABLE_USER_ROLES_APPLICANTS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "user_roles_applicants",
	}}
	tbl.APPLICANT_DATA = NewJSONField("applicant_data", tbl.TableInfo)
	tbl.APPLICANT_FORM_ID = NewNumberField("applicant_form_id", tbl.TableInfo)
	tbl.APPLICATION_ID = NewNumberField("application_id", tbl.TableInfo)
	tbl.USER_ROLE_ID = NewNumberField("user_role_id", tbl.TableInfo)
	return tbl
}

type TABLE_USER_ROLES_STUDENTS struct {
	TableInfo
	STUDENT_DATA JSONField
	TEAM_ID      NumberField
	USER_ROLE_ID NumberField
}

func USER_ROLES_STUDENTS() TABLE_USER_ROLES_STUDENTS {
	tbl := TABLE_USER_ROLES_STUDENTS{TableInfo: TableInfo{
		Schema: "public",
		Name:   "user_roles_students",
	}}
	tbl.STUDENT_DATA = NewJSONField("student_data", tbl.TableInfo)
	tbl.TEAM_ID = NewNumberField("team_id", tbl.TableInfo)
	tbl.USER_ROLE_ID = NewNumberField("user_role_id", tbl.TableInfo)
	return tbl
}

type USERS struct {
	TableInfo   `sq:"name=users"`
	DISPLAYNAME StringField `sq:"type=TEXT"`
	EMAIL       StringField `sq:"type=TEXT"`
	PASSWORD    StringField `sq:"type=TEXT"`
	USER_ID     NumberField `sq:"type=BIGINT misc=NOT_NULL,PRIMARY_KEY"`
	// DATA        JSONField   `sq:"type=JSONB"`
}

func (u USERS) Constraints(db interface{}, c interface{}) error {
	/*
		s := NEW_SESSIONS("")
		c.ForeignKey(u.USER_ID, s.USER_ID)
		c.Unique(u.DISPLAYNAME, u.EMAIL)
		c.Index(u.EMAIL) // do it like this...? how to provide method for checking if index already exists?
		if noindex {
			db.Exec("SET INDEX")
		}
		c.ChangeSet(map[Field, string, string]func() error {
			u.USER_ID,"INT","BIGINT": func() error {
				db.Exec(`ALTER TABLE ? ALTER COLUMN ? TYPE BIGINT`, u, u.USER_ID)
			},
		})
		basically you bake in migrations for specific fields+situations, using
		some code. Whenever Ensuretables cannot safely do a change, it will
		consult the changeset to see if there is some procedure registered for
		it. It will then execute that procedure. It is up to you to make the
		procedure idempotent/not block current transactions by specifying your
		own migration DDL routines, like creating a secondary table/column,
		writing to it, backfilling, dual writing (have to use triggers for
		this). This means that one server can be tasked with running the DDL
		while other servers read obliviously from it.  It is the code's
		responsibility to migrate data in such a way that does not disrupt
		existing servers reading/writing to the db (usually by modifying an
		entirely separate column that they are oblivious to). Once the two
		columns/tables are in sync, the existing servers can upgrade to the
		version of the application that reads from the new column/table
		instead. Once all servers have been migrated to this new version, the
		DDL server can begin deleting the old table/column.
	*/
	return nil
}

func NEW_USERS(alias string) USERS {
	tbl := USERS{}
	tbl.TableInfo.Alias = alias
	_ = ReflectTable(&tbl)
	return tbl
}

type VIEW_V_APPLICATIONS struct {
	TableInfo
	APPLICANT1_ANSWERS      JSONField
	APPLICANT1_DISPLAYNAME  StringField
	APPLICANT1_EMAIL        StringField
	APPLICANT1_USER_ID      NumberField
	APPLICANT1_USER_ROLE_ID NumberField
	APPLICANT2_ANSWERS      JSONField
	APPLICANT2_DISPLAYNAME  StringField
	APPLICANT2_EMAIL        StringField
	APPLICANT2_USER_ID      NumberField
	APPLICANT2_USER_ROLE_ID NumberField
	APPLICANT_FORM_ID       NumberField
	APPLICANT_QUESTIONS     JSONField
	APPLICATION_ANSWERS     JSONField
	APPLICATION_FORM_ID     NumberField
	APPLICATION_ID          NumberField
	APPLICATION_QUESTIONS   JSONField
	COHORT                  StringField
	CREATED_AT              TimeField
	CREATOR_USER_ROLE_ID    NumberField
	DELETED_AT              TimeField
	MAGICSTRING             StringField
	PROJECT_LEVEL           StringField
	STATUS                  StringField
	SUBMITTED               BooleanField
	UPDATED_AT              TimeField
}

func V_APPLICATIONS() VIEW_V_APPLICATIONS {
	tbl := VIEW_V_APPLICATIONS{TableInfo: TableInfo{
		Schema: "app",
		Name:   "v_applications",
	}}
	tbl.APPLICANT1_ANSWERS = NewJSONField("applicant1_answers", tbl.TableInfo)
	tbl.APPLICANT1_DISPLAYNAME = NewStringField("applicant1_displayname", tbl.TableInfo)
	tbl.APPLICANT1_EMAIL = NewStringField("applicant1_email", tbl.TableInfo)
	tbl.APPLICANT1_USER_ID = NewNumberField("applicant1_user_id", tbl.TableInfo)
	tbl.APPLICANT1_USER_ROLE_ID = NewNumberField("applicant1_user_role_id", tbl.TableInfo)
	tbl.APPLICANT2_ANSWERS = NewJSONField("applicant2_answers", tbl.TableInfo)
	tbl.APPLICANT2_DISPLAYNAME = NewStringField("applicant2_displayname", tbl.TableInfo)
	tbl.APPLICANT2_EMAIL = NewStringField("applicant2_email", tbl.TableInfo)
	tbl.APPLICANT2_USER_ID = NewNumberField("applicant2_user_id", tbl.TableInfo)
	tbl.APPLICANT2_USER_ROLE_ID = NewNumberField("applicant2_user_role_id", tbl.TableInfo)
	tbl.APPLICANT_FORM_ID = NewNumberField("applicant_form_id", tbl.TableInfo)
	tbl.APPLICANT_QUESTIONS = NewJSONField("applicant_questions", tbl.TableInfo)
	tbl.APPLICATION_ANSWERS = NewJSONField("application_answers", tbl.TableInfo)
	tbl.APPLICATION_FORM_ID = NewNumberField("application_form_id", tbl.TableInfo)
	tbl.APPLICATION_ID = NewNumberField("application_id", tbl.TableInfo)
	tbl.APPLICATION_QUESTIONS = NewJSONField("application_questions", tbl.TableInfo)
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.CREATED_AT = NewTimeField("created_at", tbl.TableInfo)
	tbl.CREATOR_USER_ROLE_ID = NewNumberField("creator_user_role_id", tbl.TableInfo)
	tbl.DELETED_AT = NewTimeField("deleted_at", tbl.TableInfo)
	tbl.MAGICSTRING = NewStringField("magicstring", tbl.TableInfo)
	tbl.PROJECT_LEVEL = NewStringField("project_level", tbl.TableInfo)
	tbl.STATUS = NewStringField("status", tbl.TableInfo)
	tbl.SUBMITTED = NewBooleanField("submitted", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type VIEW_V_SUBMISSIONS struct {
	TableInfo
	ANSWERS            JSONField
	COHORT             StringField
	END_AT             TimeField
	MILESTONE          StringField
	OVERRIDE_OPEN      BooleanField
	PROJECT_LEVEL      StringField
	QUESTIONS          JSONField
	START_AT           TimeField
	SUBMISSION_FORM_ID NumberField
	SUBMISSION_ID      NumberField
	SUBMITTED          BooleanField
	TEAM_ID            NumberField
	TEAM_NAME          StringField
	UPDATED_AT         TimeField
}

func V_SUBMISSIONS() VIEW_V_SUBMISSIONS {
	tbl := VIEW_V_SUBMISSIONS{TableInfo: TableInfo{
		Schema: "app",
		Name:   "v_submissions",
	}}
	tbl.ANSWERS = NewJSONField("answers", tbl.TableInfo)
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.END_AT = NewTimeField("end_at", tbl.TableInfo)
	tbl.MILESTONE = NewStringField("milestone", tbl.TableInfo)
	tbl.OVERRIDE_OPEN = NewBooleanField("override_open", tbl.TableInfo)
	tbl.PROJECT_LEVEL = NewStringField("project_level", tbl.TableInfo)
	tbl.QUESTIONS = NewJSONField("questions", tbl.TableInfo)
	tbl.START_AT = NewTimeField("start_at", tbl.TableInfo)
	tbl.SUBMISSION_FORM_ID = NewNumberField("submission_form_id", tbl.TableInfo)
	tbl.SUBMISSION_ID = NewNumberField("submission_id", tbl.TableInfo)
	tbl.SUBMITTED = NewBooleanField("submitted", tbl.TableInfo)
	tbl.TEAM_ID = NewNumberField("team_id", tbl.TableInfo)
	tbl.TEAM_NAME = NewStringField("team_name", tbl.TableInfo)
	tbl.UPDATED_AT = NewTimeField("updated_at", tbl.TableInfo)
	return tbl
}

type VIEW_V_TEAM_EVALUATIONS struct {
	TableInfo
	COHORT                   StringField
	EVALUATEE_PROJECT_LEVEL  StringField
	EVALUATEE_TEAM_ID        NumberField
	EVALUATEE_TEAM_NAME      StringField
	EVALUATION_ANSWERS       JSONField
	EVALUATION_END_AT        TimeField
	EVALUATION_FORM_ID       NumberField
	EVALUATION_OVERRIDE_OPEN BooleanField
	EVALUATION_QUESTIONS     JSONField
	EVALUATION_START_AT      TimeField
	EVALUATION_SUBMITTED     BooleanField
	EVALUATION_UPDATED_AT    TimeField
	EVALUATOR_PROJECT_LEVEL  StringField
	EVALUATOR_TEAM_ID        NumberField
	EVALUATOR_TEAM_NAME      StringField
	MILESTONE                StringField
	STAGE                    StringField
	SUBMISSION_ANSWERS       JSONField
	SUBMISSION_END_AT        TimeField
	SUBMISSION_FORM_ID       NumberField
	SUBMISSION_ID            NumberField
	SUBMISSION_OVERRIDE_OPEN BooleanField
	SUBMISSION_QUESTIONS     JSONField
	SUBMISSION_START_AT      TimeField
	SUBMISSION_SUBMITTED     BooleanField
	SUBMISSION_UPDATED_AT    TimeField
	TEAM_EVALUATION_ID       NumberField
}

func V_TEAM_EVALUATIONS() VIEW_V_TEAM_EVALUATIONS {
	tbl := VIEW_V_TEAM_EVALUATIONS{TableInfo: TableInfo{
		Schema: "app",
		Name:   "v_team_evaluations",
	}}
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.EVALUATEE_PROJECT_LEVEL = NewStringField("evaluatee_project_level", tbl.TableInfo)
	tbl.EVALUATEE_TEAM_ID = NewNumberField("evaluatee_team_id", tbl.TableInfo)
	tbl.EVALUATEE_TEAM_NAME = NewStringField("evaluatee_team_name", tbl.TableInfo)
	tbl.EVALUATION_ANSWERS = NewJSONField("evaluation_answers", tbl.TableInfo)
	tbl.EVALUATION_END_AT = NewTimeField("evaluation_end_at", tbl.TableInfo)
	tbl.EVALUATION_FORM_ID = NewNumberField("evaluation_form_id", tbl.TableInfo)
	tbl.EVALUATION_OVERRIDE_OPEN = NewBooleanField("evaluation_override_open", tbl.TableInfo)
	tbl.EVALUATION_QUESTIONS = NewJSONField("evaluation_questions", tbl.TableInfo)
	tbl.EVALUATION_START_AT = NewTimeField("evaluation_start_at", tbl.TableInfo)
	tbl.EVALUATION_SUBMITTED = NewBooleanField("evaluation_submitted", tbl.TableInfo)
	tbl.EVALUATION_UPDATED_AT = NewTimeField("evaluation_updated_at", tbl.TableInfo)
	tbl.EVALUATOR_PROJECT_LEVEL = NewStringField("evaluator_project_level", tbl.TableInfo)
	tbl.EVALUATOR_TEAM_ID = NewNumberField("evaluator_team_id", tbl.TableInfo)
	tbl.EVALUATOR_TEAM_NAME = NewStringField("evaluator_team_name", tbl.TableInfo)
	tbl.MILESTONE = NewStringField("milestone", tbl.TableInfo)
	tbl.STAGE = NewStringField("stage", tbl.TableInfo)
	tbl.SUBMISSION_ANSWERS = NewJSONField("submission_answers", tbl.TableInfo)
	tbl.SUBMISSION_END_AT = NewTimeField("submission_end_at", tbl.TableInfo)
	tbl.SUBMISSION_FORM_ID = NewNumberField("submission_form_id", tbl.TableInfo)
	tbl.SUBMISSION_ID = NewNumberField("submission_id", tbl.TableInfo)
	tbl.SUBMISSION_OVERRIDE_OPEN = NewBooleanField("submission_override_open", tbl.TableInfo)
	tbl.SUBMISSION_QUESTIONS = NewJSONField("submission_questions", tbl.TableInfo)
	tbl.SUBMISSION_START_AT = NewTimeField("submission_start_at", tbl.TableInfo)
	tbl.SUBMISSION_SUBMITTED = NewBooleanField("submission_submitted", tbl.TableInfo)
	tbl.SUBMISSION_UPDATED_AT = NewTimeField("submission_updated_at", tbl.TableInfo)
	tbl.TEAM_EVALUATION_ID = NewNumberField("team_evaluation_id", tbl.TableInfo)
	return tbl
}

type VIEW_V_TEAMS struct {
	TableInfo
	ADVISER_DISPLAYNAME   StringField
	ADVISER_EMAIL         StringField
	ADVISER_USER_ID       NumberField
	ADVISER_USER_ROLE_ID  NumberField
	COHORT                StringField
	MENTOR_DISPLAYNAME    StringField
	MENTOR_EMAIL          StringField
	MENTOR_USER_ID        NumberField
	MENTOR_USER_ROLE_ID   NumberField
	PROJECT_LEVEL         StringField
	STATUS                StringField
	STUDENT1_DISPLAYNAME  StringField
	STUDENT1_EMAIL        StringField
	STUDENT1_USER_ID      NumberField
	STUDENT1_USER_ROLE_ID NumberField
	STUDENT2_DISPLAYNAME  StringField
	STUDENT2_EMAIL        StringField
	STUDENT2_USER_ID      NumberField
	STUDENT2_USER_ROLE_ID NumberField
	TEAM_ID               NumberField
	TEAM_NAME             StringField
}

func V_TEAMS() VIEW_V_TEAMS {
	tbl := VIEW_V_TEAMS{TableInfo: TableInfo{
		Schema: "app",
		Name:   "v_teams",
	}}
	tbl.ADVISER_DISPLAYNAME = NewStringField("adviser_displayname", tbl.TableInfo)
	tbl.ADVISER_EMAIL = NewStringField("adviser_email", tbl.TableInfo)
	tbl.ADVISER_USER_ID = NewNumberField("adviser_user_id", tbl.TableInfo)
	tbl.ADVISER_USER_ROLE_ID = NewNumberField("adviser_user_role_id", tbl.TableInfo)
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.MENTOR_DISPLAYNAME = NewStringField("mentor_displayname", tbl.TableInfo)
	tbl.MENTOR_EMAIL = NewStringField("mentor_email", tbl.TableInfo)
	tbl.MENTOR_USER_ID = NewNumberField("mentor_user_id", tbl.TableInfo)
	tbl.MENTOR_USER_ROLE_ID = NewNumberField("mentor_user_role_id", tbl.TableInfo)
	tbl.PROJECT_LEVEL = NewStringField("project_level", tbl.TableInfo)
	tbl.STATUS = NewStringField("status", tbl.TableInfo)
	tbl.STUDENT1_DISPLAYNAME = NewStringField("student1_displayname", tbl.TableInfo)
	tbl.STUDENT1_EMAIL = NewStringField("student1_email", tbl.TableInfo)
	tbl.STUDENT1_USER_ID = NewNumberField("student1_user_id", tbl.TableInfo)
	tbl.STUDENT1_USER_ROLE_ID = NewNumberField("student1_user_role_id", tbl.TableInfo)
	tbl.STUDENT2_DISPLAYNAME = NewStringField("student2_displayname", tbl.TableInfo)
	tbl.STUDENT2_EMAIL = NewStringField("student2_email", tbl.TableInfo)
	tbl.STUDENT2_USER_ID = NewNumberField("student2_user_id", tbl.TableInfo)
	tbl.STUDENT2_USER_ROLE_ID = NewNumberField("student2_user_role_id", tbl.TableInfo)
	tbl.TEAM_ID = NewNumberField("team_id", tbl.TableInfo)
	tbl.TEAM_NAME = NewStringField("team_name", tbl.TableInfo)
	return tbl
}

type VIEW_V_TEAMS_AND_STUDENTS struct {
	TableInfo
	ADVISER_USER_ROLE_ID NumberField
	MENTOR_USER_ROLE_ID  NumberField
	PROJECT_LEVEL        StringField
	STUDENT1_DISPLAYNAME StringField
	STUDENT2_DISPLAYNAME StringField
	TEAM_ID              NumberField
	TEAM_NAME            StringField
}

func V_TEAMS_AND_STUDENTS() VIEW_V_TEAMS_AND_STUDENTS {
	tbl := VIEW_V_TEAMS_AND_STUDENTS{TableInfo: TableInfo{
		Schema: "app",
		Name:   "v_teams_and_students",
	}}
	tbl.ADVISER_USER_ROLE_ID = NewNumberField("adviser_user_role_id", tbl.TableInfo)
	tbl.MENTOR_USER_ROLE_ID = NewNumberField("mentor_user_role_id", tbl.TableInfo)
	tbl.PROJECT_LEVEL = NewStringField("project_level", tbl.TableInfo)
	tbl.STUDENT1_DISPLAYNAME = NewStringField("student1_displayname", tbl.TableInfo)
	tbl.STUDENT2_DISPLAYNAME = NewStringField("student2_displayname", tbl.TableInfo)
	tbl.TEAM_ID = NewNumberField("team_id", tbl.TableInfo)
	tbl.TEAM_NAME = NewStringField("team_name", tbl.TableInfo)
	return tbl
}

type VIEW_V_USER_EVALUATIONS struct {
	TableInfo
	COHORT                   StringField
	EVALUATEE_PROJECT_LEVEL  StringField
	EVALUATEE_TEAM_ID        NumberField
	EVALUATEE_TEAM_NAME      StringField
	EVALUATION_ANSWERS       JSONField
	EVALUATION_END_AT        TimeField
	EVALUATION_FORM_ID       NumberField
	EVALUATION_OVERRIDE_OPEN BooleanField
	EVALUATION_QUESTIONS     JSONField
	EVALUATION_START_AT      TimeField
	EVALUATION_SUBMITTED     BooleanField
	EVALUATION_UPDATED_AT    TimeField
	EVALUATOR_DISPLAYNAME    StringField
	EVALUATOR_ROLE           StringField
	EVALUATOR_USER_ID        NumberField
	EVALUATOR_USER_ROLE_ID   NumberField
	MILESTONE                StringField
	STAGE                    StringField
	SUBMISSION_ANSWERS       JSONField
	SUBMISSION_END_AT        TimeField
	SUBMISSION_FORM_ID       NumberField
	SUBMISSION_ID            NumberField
	SUBMISSION_OVERRIDE_OPEN BooleanField
	SUBMISSION_QUESTIONS     JSONField
	SUBMISSION_START_AT      TimeField
	SUBMISSION_SUBMITTED     BooleanField
	SUBMISSION_UPDATED_AT    TimeField
	USER_EVALUATION_ID       NumberField
}

func V_USER_EVALUATIONS() VIEW_V_USER_EVALUATIONS {
	tbl := VIEW_V_USER_EVALUATIONS{TableInfo: TableInfo{
		Schema: "app",
		Name:   "v_user_evaluations",
	}}
	tbl.COHORT = NewStringField("cohort", tbl.TableInfo)
	tbl.EVALUATEE_PROJECT_LEVEL = NewStringField("evaluatee_project_level", tbl.TableInfo)
	tbl.EVALUATEE_TEAM_ID = NewNumberField("evaluatee_team_id", tbl.TableInfo)
	tbl.EVALUATEE_TEAM_NAME = NewStringField("evaluatee_team_name", tbl.TableInfo)
	tbl.EVALUATION_ANSWERS = NewJSONField("evaluation_answers", tbl.TableInfo)
	tbl.EVALUATION_END_AT = NewTimeField("evaluation_end_at", tbl.TableInfo)
	tbl.EVALUATION_FORM_ID = NewNumberField("evaluation_form_id", tbl.TableInfo)
	tbl.EVALUATION_OVERRIDE_OPEN = NewBooleanField("evaluation_override_open", tbl.TableInfo)
	tbl.EVALUATION_QUESTIONS = NewJSONField("evaluation_questions", tbl.TableInfo)
	tbl.EVALUATION_START_AT = NewTimeField("evaluation_start_at", tbl.TableInfo)
	tbl.EVALUATION_SUBMITTED = NewBooleanField("evaluation_submitted", tbl.TableInfo)
	tbl.EVALUATION_UPDATED_AT = NewTimeField("evaluation_updated_at", tbl.TableInfo)
	tbl.EVALUATOR_DISPLAYNAME = NewStringField("evaluator_displayname", tbl.TableInfo)
	tbl.EVALUATOR_ROLE = NewStringField("evaluator_role", tbl.TableInfo)
	tbl.EVALUATOR_USER_ID = NewNumberField("evaluator_user_id", tbl.TableInfo)
	tbl.EVALUATOR_USER_ROLE_ID = NewNumberField("evaluator_user_role_id", tbl.TableInfo)
	tbl.MILESTONE = NewStringField("milestone", tbl.TableInfo)
	tbl.STAGE = NewStringField("stage", tbl.TableInfo)
	tbl.SUBMISSION_ANSWERS = NewJSONField("submission_answers", tbl.TableInfo)
	tbl.SUBMISSION_END_AT = NewTimeField("submission_end_at", tbl.TableInfo)
	tbl.SUBMISSION_FORM_ID = NewNumberField("submission_form_id", tbl.TableInfo)
	tbl.SUBMISSION_ID = NewNumberField("submission_id", tbl.TableInfo)
	tbl.SUBMISSION_OVERRIDE_OPEN = NewBooleanField("submission_override_open", tbl.TableInfo)
	tbl.SUBMISSION_QUESTIONS = NewJSONField("submission_questions", tbl.TableInfo)
	tbl.SUBMISSION_START_AT = NewTimeField("submission_start_at", tbl.TableInfo)
	tbl.SUBMISSION_SUBMITTED = NewBooleanField("submission_submitted", tbl.TableInfo)
	tbl.SUBMISSION_UPDATED_AT = NewTimeField("submission_updated_at", tbl.TableInfo)
	tbl.USER_EVALUATION_ID = NewNumberField("user_evaluation_id", tbl.TableInfo)
	return tbl
}

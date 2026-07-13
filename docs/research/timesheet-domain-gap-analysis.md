# From timesheet prefill to timesheet: domain and business-logic gap analysis

_Research date: 2026-07-12_

## Question and scope

What would Shiet need to become a timesheet-like application rather than a tool that prepares data for another timesheet?

This note compares the implemented application and its documented intent with primary sources: labour authorities and legislation, official accounting/payroll APIs, official enterprise-timesheet documentation, and first-party privacy guidance. It identifies product/domain requirements; it is not legal advice. Wage-and-hour rules vary by jurisdiction, worker classification, contract, and collective agreement, so the product should model policies rather than hard-code one legal interpretation.

## Executive conclusion

Shiet is already unusually close to a **personal timesheet workbench**: it has periods, timezone-aware intervals, calendar imports, manual gap fills, categorization, explicit conflict review, descriptions, summaries, and configurable exports. It is not yet a reliable **timesheet ledger** or an **employer system of record**.

The first change should not be “add an approval button.” It should be to introduce a canonical, user-attested `TimeEntry` that is separate from imported calendar events and other evidence. Today, active calendar events flow directly into exports, including unassigned events; all-day events become 24-hour export entries; overlaps can be double-counted; and the overlap decision is not applied by the export builder. Imported observations are therefore being treated as accounting truth before the user attests them. A real timesheet should make calendar events, GitHub activity, and Slack activity **provenance/evidence that can seed or explain entries**, while `TimeEntry` is the sole source for totals, submission, approval, and downstream integrations.

After that boundary exists, the minimum viable product is a **single-user, authoritative personal timesheet**: normalized entries, a backend-owned work schedule, independent allocation/pay dimensions, completeness and validation rules, finalization/versioned amendments, and deterministic export delivery. Becoming an **organizational timesheet system** is a second product boundary. It requires employee and organization identity, approver authority, shared storage, permissions, notifications, retention administration, and cross-user audit—capabilities that contradict the current “one local install, one independent user, no accounts/admin” shape in [DESIGN.md](../../DESIGN.md).

## What “timesheet” means in this analysis

There are three related products that should not be conflated:

1. **Attendance/time-and-attendance record** — measures when a person worked, took breaks, or took leave, usually for wage-and-hour purposes.
2. **Allocative/project timesheet** — distributes worked time across projects, customers, tasks, cost objectives, or indirect accounts for costing and billing.
3. **Submitted organizational record** — an employee attests a period, an authorized party reviews it, and a frozen/versioned result feeds payroll, accounting, billing, or compliance evidence.

Primary sources show why all three dimensions matter:

- The U.S. Department of Labor says covered employers must retain accurate employee identity/pay data, hours worked each day, total hours each workweek, straight-time and overtime earnings, deductions, wages paid, and the pay period; it permits any timekeeping method only if it is complete and accurate. It also distinguishes a pay period from the fixed 168-hour workweek used to compute overtime and prohibits averaging across weeks. ([DOL recordkeeping](https://www.dol.gov/general/topic/workhours/hoursrecordkeeping), [DOL Fact Sheet #21](https://www.dol.gov/agencies/whd/fact-sheets/21-flsa-recordkeeping), [DOL overtime](https://www.dol.gov/index.php/agencies/whd/overtime))
- Canadian federal rules likewise require hours worked each day and retain records for at least three years, alongside wage basis, overtime, vacation, holiday, and leave pay information. Federal standard hours are expressed daily and weekly, with special rules for approved averaging and general holidays. ([Canada Labour Standards Regulations, s. 24](https://laws-lois.justice.gc.ca/eng/regulations/C.R.C.%2C_c._986/section-24.html), [Canada Labour Code, s. 169](https://laws-lois.justice.gc.ca/eng/acts/L-2/section-169.html?wbdisable=true), [Canada Labour Code, s. 252](https://laws-lois.justice.gc.ca/eng/acts/L-2/section-252.html?txthl=kept))
- The EU Working Time Directive treats daily/weekly rest, breaks, annual leave, and maximum weekly work as distinct concerns. The Court of Justice held that Member States must require a system capable of measuring each worker's daily working time; its reasoning emphasizes objective and reliable data. ([Directive 2003/88/EC](https://eur-lex.europa.eu/legal-content/EN/TXT/?qid=1561450242568&uri=CELEX%3A32003L0088), [CJEU C-55/18](https://infocuria.curia.europa.eu/tabs/redirect/juris/document/document.jsf?cid=7740630&dir=&docid=214043&doclang=EN&mode=req&occ=first&pageIndex=0&part=1&text=))
- Accounting/payroll interfaces add allocation semantics. Xero's payroll timesheet model identifies employee, payroll calendar, period, status, earnings rate, daily units, and optional tracking item; approved timesheets enter a pay run and later become processed. DCAA guidance for U.S. government contractors adds labor distribution by cost objective, traceability to work authorization, employee certification, supervisor approval, and reconciliation to payroll/cost ledgers. ([Xero Payroll AU integration guide](https://developer.xero.com/documentation/api/payrollau/integration-guide), [Xero Payroll UK timesheets](https://developer.xero.com/documentation/api/payrolluk/timesheets), [DCAA Contract Audit Manual, chapter 5](https://www.dcaa.mil/Portals/88/Documents/Guidance/CAM/CAM%20Chapter%2005%20DFARS%20Business%20Sys.pdf?ver=v255V0QIJTaCciMaUQBSBg%3D%3D))

These are examples of the policy space, not a claim that every Shiet user is covered by every regime.

## Current Shiet baseline

### Implemented strengths

- `period` and `tz_segment` provide deterministic date ranges and timezone-aware placement.
- Imported event facts are separated from durable category overlays, so resync can preserve user decisions.
- `gap_fill` supports user-created and edited intervals with notes/descriptions.
- `review_item` and the review-policy service make ambiguous sync changes explicit rather than silently discarding work.
- Categories, category memory, selectable calendars, descriptions, activity evidence, custom export templates, and CSV/TSV/text output already support the preparation loop well.
- The database contains a `submission` table designed for versioned immutable snapshots.

These are visible in [the initial migration](../../internal/db/migrations/00001_init.sql), [service domain types](../../internal/service/types.go), [the application API](../../proto/shiet/app/v1/application.proto), and [export services](../../internal/service/export.go).

### Documented intent that is not implemented

`DESIGN.md` describes finalization as an immutable, versioned submission. The live repository only has schema/sqlc queries for submissions. There is no submission/finalize service or API, and the sidebar explicitly disables “Finalize period.” See [submission queries](../../internal/db/query/submission.sql) and [ScheduleSidebar.tsx](../../frontend/src/components/schedule/ScheduleSidebar.tsx).

### Implemented correctness gaps that matter once exports become authoritative

- **Imported evidence is export truth.** `BuildPeriodExport` combines active events and gap fills directly. `eventToExportEntry` emits an event even without a category decision.
- **All-day events become 1,440 minutes.** Export constants define a day as `00:00–24:00`, and all-day events use that full span.
- **Overlap decisions do not govern totals.** Export reads category overlays but not `resolved_overlap`; overlapping active events can therefore double-count.
- **Cross-midnight durations are clipped, not split.** The event/fill export conversion caps an entry at the end of its start day. Very short or malformed converted spans are forced to at least 15 minutes.
- **Every calendar date receives the daily target.** `target_minutes` is `target_hours_per_day × number of dates`, including weekends and holidays. There is no backend work-calendar, holiday, leave, part-time, or schedule-exception model.
- **Manual-entry validation is structural only.** It checks period/day bounds and that start precedes end. It does not enforce overlap, required allocation, locked periods, maximum hours, rounding, breaks, or policy-specific completeness.
- **Category is the only allocation dimension.** There is no independent client, project, task/work item, cost centre/cost objective, billable status, pay/earnings code, leave code, rate, or work authorization.

The relevant implementation is [manual_event.go](../../internal/service/manual_event.go) and [export.go](../../internal/service/export.go). These gaps are tolerable for a draft helper; they are not acceptable invariants for a record used to pay people or bill customers.

## Required domain model

### 1. Canonical time ledger

Introduce a `TimeEntry` as the only unit included in authoritative totals and submissions:

```text
TimeEntry
  id, owner_id, period_id
  local_work_date, start_instant?, end_instant?, duration_minutes
  work_type: worked | paid_leave | unpaid_leave | holiday | break | adjustment
  project_id?, task_id?, cost_objective_id?, pay_code_id?
  billable_status?, customer_id?, service_id?
  description
  provenance[]
  attestation_state, created_by, created_at, updated_by, updated_at
```

Intervals should be supported because Shiet's strongest UX is timeline-based, but the ledger should also permit duration-only daily entries: official payroll APIs commonly exchange daily units rather than clock intervals. Xero, for example, represents timesheet lines by date, earnings rate, optional tracking item, and units. ([Xero Payroll UK timesheets](https://developer.xero.com/documentation/api/payrolluk/timesheets))

`Provenance` should link an entry to imported event/evidence IDs, suggestion method, and source revision without making the source record itself payable. The creation flow becomes `observation -> proposal -> user confirmation -> TimeEntry`. User-authored entries can have no external provenance.

Core invariants:

- Duration is positive and uses exact instants or an explicit duration; overnight work is split or attributed by a declared policy, never silently clipped.
- A source observation can support zero, one, or several entries; an entry can cite several observations.
- All-day calendar events are non-time evidence unless the user/policy explicitly converts them.
- Deleted/changed source observations cannot mutate a submitted snapshot; they can create a review proposal for the live draft.
- Overlap rules apply to ledger entries, not merely visual events. Double-counting must be explicitly allowed (for example, parallel billable allocation) or blocked/resolved.
- Entries used in a submission retain the effective allocation and policy labels even if master data is later renamed or deactivated.

### 2. Independent dimensions, not a larger `Category`

Keep category as a personal classification/tag, but add independent references:

- **Worker/employment:** employee or contractor identity, employment/classification, work location/jurisdiction, effective schedule, and applicable policy set.
- **Customer/project/task:** who benefits from the work and what assignment/work item authorized it.
- **Costing:** cost objective or cost centre, direct/indirect classification.
- **Billing:** billable/non-billable/billed status, customer and service/item references, optional billing rate (ideally rates stay in the destination accounting system unless Shiet must calculate money).
- **Payroll:** earning/pay code (regular, overtime, premium, paid leave, holiday), with rate references rather than prematurely calculating wages.

This separation is demonstrated by first-party downstream contracts. Xero keeps employee, payroll calendar, earnings rate, and tracking item distinct. QuickBooks' official desktop time-tracking contract requires customer and service references when time is billable, and separately tracks billable/billed state. ([Xero Payroll AU timesheets](https://developer.xero.com/documentation/api/payrollau/timesheets), [QuickBooks TimeTrackingAdd](https://developer.intuit.com/app/developer/qbdesktop/docs/api-reference/qbdesktop/timetrackingadd))

Master records need stable IDs, active/inactive state, effective dates, and mapping to external-system IDs. Historical entries must remain readable when a project or code is deactivated.

### 3. Work schedule and expected-time policy

Create one backend-owned `WorkSchedule`/`ExpectedTimePolicy` used by the scheduler, gap computation, completeness checks, stats, and exports:

```text
WorkSchedule: timezone, workweek_start, effective dates
ScheduleDay: weekday, expected_minutes, working windows
ScheduleException: date, holiday | leave | changed_hours, expected_minutes
```

Pay period and workweek are separate concepts. A biweekly/semi-monthly pay period can contain multiple regulatory workweeks; overtime/completeness must be computed on the applicable workweek, not averaged over the pay period. ([DOL overtime](https://www.dol.gov/index.php/agencies/whd/overtime))

Leave, holidays, and breaks must be first-class classifications or schedule exceptions—not calendar categories. Microsoft Graph's first-party workforce model likewise separates time cards (including clock events and breaks) from time off and time-off reasons. ([Microsoft Graph `timeCard`](https://learn.microsoft.com/en-us/graph/api/resources/timecard?view=graph-rest-1.0), [Microsoft Graph `schedule`](https://learn.microsoft.com/en-us/graph/api/resources/schedule))

### 4. Versioned policy engine

Avoid claiming universal “compliance.” Implement configurable, versioned policies selected by organization/worker/jurisdiction:

- required fields and allowed projects/pay codes;
- expected hours by day/workweek;
- overlapping-entry policy;
- rounding increment/mode and whether rounding is permitted;
- paid/unpaid break treatment and missed-break attestations;
- overtime thresholds and averaging arrangements;
- maximum/minimum hours and rest warnings;
- submission deadline, grace period, lock behavior, and correction rules.

Store the effective policy version or a frozen policy snapshot with each submission. This is necessary because results must remain explainable after a schedule or policy changes. Tempo's official approval log similarly preserves the workload/holiday scheme that was effective at submission. ([Tempo approval documentation](https://help.tempo.io/timesheets/latest/approving-or-rejecting-timesheets))

Do not silently infer break/pay treatment. U.S. federal guidance, for example, generally counts short breaks as work but treats bona fide meal periods differently; state rules can be more protective. ([DOL breaks and meal periods](https://www.dol.gov/general/topic/workhours/breaks), [DOL Hours Worked Advisor](https://webapps.dol.gov/elaws/whd/flsa/hoursworked/screenEr4.asp))

## Workflow and state machines

### Personal authoritative-timesheet MVP

```text
OPEN DRAFT
  -> READY (all blocking reviews resolved; required allocation complete)
  -> FINALIZED vN (immutable snapshot + attestation + policy snapshot)
  -> EXPORTED/DELIVERED (per destination receipt; does not mutate snapshot)
  -> AMENDMENT DRAFT
  -> FINALIZED vN+1 (supersedes vN; reason required)
```

Finalization should be a transaction that:

1. recomputes totals from canonical entries;
2. runs blocking validation and records warnings/overrides;
3. requires an attestation (actor, timestamp, statement/version);
4. freezes entries, allocation display values, expected-time/policy snapshot, totals, and source/provenance references;
5. creates a new version rather than overwriting history.

An export is a delivery artifact, not the submission itself. Track destination, template/schema version, checksum, external identifier, attempt/result timestamps, and whether the destination acknowledged it. This prevents a successful payroll push from being accidentally repeated and supports reconciliation.

### Organizational system-of-record extension

```text
OPEN -> SUBMITTED -> APPROVED -> CLOSED/PROCESSED
          |             |
          v             v
        REJECTED      REOPENED
          |             |
          +--> OPEN/AMENDMENT <--+
```

Required transitions and invariants:

- Submitter and approver are authenticated identities; self-approval is forbidden unless an explicit organization policy allows it.
- Submission locks employee editing. Rejection/reopen includes actor, timestamp, reason, and returns a controlled editable version.
- Approval is against a specific immutable version. Editing material fields invalidates approval; correction produces a new version.
- Period close/process is distinct from approval. A processed payroll/billing record is corrected through adjustment/amendment, not mutable history.
- Approval may be whole-timesheet, per-project, or both. If both exist, their states must be independent and sequencing explicit.

Official products expose these same distinctions. Tempo documents Open -> Waiting for Approval -> Approved/Rejected, prevents edits after submission, prevents self-approval, and logs submit/approve/reject actions with actor/time. It also distinguishes whole-timesheet approval from project-time approval. Xero distinguishes Draft, Approved, and Processed payroll timesheets. Oracle documents separate project-time approval authority and that approved time can feed payroll and customer billing. ([Tempo timesheet approvals](https://help.tempo.io/timesheets/latest/timesheet-approvals), [Tempo project approvals](https://help.tempo.io/timesheets/latest/project-approvals), [Xero Payroll AU integration guide](https://developer.xero.com/documentation/api/payrollau/integration-guide), [Oracle NetSuite time approval](https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_N1192683.html))

The local desktop architecture can implement the personal state machine. It cannot prove independent manager approval merely by adding a local approver name: organizational approval requires a shared authority boundary and authenticated actor.

## Audit, retention, privacy, and integration implications

### Audit and corrections

Use an append-only domain event/audit log for consequential actions: entry create/update/delete, source conversion, policy override, submit/finalize, approve/reject/reopen, export, destination acknowledgement, and amendment. Record actor, time, object/version, action, reason, and before/after hashes or changed fields. This is separate from application debug logs.

Retain immutable submission versions and explicit supersession links. Never present a regenerated report from today's mutable names/policy as the historical submission. DCAA guidance illustrates the stronger end of this requirement: daily entry, cost-objective authorization, employee certification, supervisor approval, traceability, and reconciliation. ([DCAA Information for Contractors](https://www.dcaa.mil/Portals/88/Documents/Guidance/CAM/Information%20For%20Contractors%20DCAAM%207641_90.pdf), [DCAA Contract Audit Manual, chapter 5](https://www.dcaa.mil/Portals/88/Documents/Guidance/CAM/CAM%20Chapter%2005%20DFARS%20Business%20Sys.pdf?ver=v255V0QIJTaCciMaUQBSBg%3D%3D))

Retention must be policy-driven by record type and jurisdiction, with legal-hold/export/delete administration in an organizational product. U.S. FLSA guidance distinguishes three-year payroll records from two-year supporting wage-computation records; Canadian federal rules generally require at least three years after the work. ([DOL Fact Sheet #21](https://www.dol.gov/agencies/whd/fact-sheets/21-flsa-recordkeeping), [Canada Labour Standards Regulations, s. 24](https://laws-lois.justice.gc.ca/eng/regulations/C.R.C.%2C_c._986/section-24.html))

### Privacy and access

The current local-first design reduces centralized exposure and should remain a product advantage for the personal edition. An organizational edition would process employee, calendar, project, location, and potentially leave data. It therefore needs purpose limitation, field minimization, role-based access, encryption, tenant isolation, data-subject/export workflows, configured retention, and auditable administrative access. EU Commission guidance summarizes the controlling principles as purpose limitation, minimization, accuracy, storage limitation, and integrity/confidentiality. ([European Commission GDPR principles](https://commission.europa.eu/law/law-topic/data-protection/rules-business-and-organisations/principles-gdpr/overview-principles/what-data-can-we-process-and-under-which-conditions_en))

Do not copy payroll identifiers, wage rates, or sensitive leave details into Shiet unless a supported workflow truly requires them. Prefer stable external references and purpose-specific labels.

### Export/integration contracts

Each destination adapter needs:

- organization/worker and master-data mapping;
- explicit target schema/version and unit/rounding rules;
- preflight validation against active destination references;
- idempotency/delivery identity and retry semantics;
- status mapping (draft/approved/processed/billed);
- response/error persistence and reconciliation;
- a rule for whether Shiet or the destination owns overtime/rate/money calculations.

Recommended default: Shiet owns accurate time, allocation, attestation, and policy evidence; payroll/accounting owns money calculations unless the integration contract explicitly requires Shiet to calculate them.

## Prioritized gap matrix

| Priority | Capability | Current state | Required outcome |
|---|---|---|---|
| P0 | Canonical time ledger | Events and gap fills are separate export inputs | `TimeEntry` is sole authoritative unit; imports are evidence/provenance |
| P0 | Accurate interval accounting | All-day=24h, overnight clipping, unresolved overlap double-count risk | Exact/split durations; all-day non-time by default; explicit overlap invariant |
| P0 | Work schedule | One target-hours/day value applied to every date | Effective workweek, weekday windows, holidays, leave, exceptions; one backend authority |
| P0 | Completeness/validation | Structural manual-entry checks and sync review | Required fields, allocation, overlap, expected-time, unresolved-review, and lock checks |
| P0 | Finalize/amend | Schema intent only; UI disabled | Transactional immutable version, attestation, validation record, superseding amendment |
| P1 | Allocation model | One free-form category | Separate worker, project/customer/task, cost, billable, pay/leave dimensions |
| P1 | Export delivery | Render/save files | Destination mapping, schema version, checksum/idempotency, receipt and reconciliation |
| P1 | Audit | Timestamps plus review decisions | Append-only actor/action/reason/version audit log |
| P1 | Policy model | Mostly implicit behavior | Versioned work/break/overtime/rounding/submission policies and overrides |
| P2 | Organizational approval | No identities, org, shared authority | Authenticated submitter/approver, permissions, reject/reopen, notifications, close/process |
| P2 | Administration | Local independent user | Org master data, policy assignment, retention, audit access, tenant boundary |
| P3 | Financial calculation | Hours only | Only if strategically chosen: rates, premiums, billing/payroll calculations and reconciliation |

## Recommended product sequence

1. **Define the trust boundary.** Add canonical `TimeEntry` + provenance and migrate export/stats/gaps to read it. Calendar events become proposals until confirmed. This fixes the highest-risk semantics before adding workflow.
2. **Make personal timesheets correct.** Add backend `WorkSchedule`, holidays/leave/break types, exact duration handling, overlap rules, allocation requirements, and a unified completeness result.
3. **Make them attestable.** Implement finalization from the existing submission intent, with immutable snapshots, attestation, validation/policy snapshots, audit events, and amendment versions.
4. **Make them interoperable.** Add independent project/customer/pay-code dimensions and one destination adapter with mapping, preflight, idempotent delivery, and reconciliation. Let the chosen destination contract shape the minimal fields.
5. **Choose the business boundary explicitly.** Either remain a privacy-first personal system that creates trustworthy submissions for external payroll/accounting, or fund the much larger organizational system: identity, shared authority, approval, admin, retention, and hosted sync. Do not simulate organizational approval locally.

## Essential versus optional

For the phrase “timesheet-like application,” the essential set is: canonical attested entries, accurate daily/workweek totals, work schedule/exceptions, independent allocation/pay classifications, completeness rules, immutable finalization/amendments, and deterministic export. Approval is essential only if Shiet claims to be an employer/organization system of record.

Timers, geofencing, screenshots, invoicing, payroll calculation, capacity planning, expenses, resource forecasting, multi-level approval, and jurisdiction-specific compliance packs are optional verticals. They should not precede the canonical ledger and policy/submission model.

The strategic opportunity is therefore not to rebuild a generic clock. It is to retain Shiet's differentiator—turning messy personal evidence into a reviewed schedule—then make the result a trustworthy, portable time ledger.

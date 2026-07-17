-- +goose Up
-- Retire flat hoursPerDay from the builtin text_summary template.
-- Per-day / period targets now come from ExpectedTime (sum of resolver minutes).
UPDATE export_template
SET body = 'Period: {{.PeriodLabel}}
{{.StartDate}} to {{.EndDate}}

Target: {{duration .TargetMinutes}}
Actual: {{duration .ActualMinutes}}
Variance: {{signedDuration .VarianceMinutes}}

Totals by category:
{{range .PeriodTotals}}  {{.Category.Name}}: {{duration .Minutes}}
{{end}}
Daily breakdown:
{{range .DailyTotals}}{{.Date}} — {{duration .ActualMinutes}} / {{duration .TargetMinutes}} target
{{if .Categories}}{{range .Categories}}  {{.Category.Name}}: {{duration .Minutes}}
{{end}}{{else}}  (no tracked time)
{{end}}
{{end}}'
WHERE key = 'text_summary' AND builtin = 1;

-- +goose Down
UPDATE export_template
SET body = 'Period: {{.PeriodLabel}}
{{.StartDate}} to {{.EndDate}}

Target: {{duration .TargetMinutes}} ({{hoursPerDay .TargetHoursPerDay}}h/day)
Actual: {{duration .ActualMinutes}}
Variance: {{signedDuration .VarianceMinutes}}

Totals by category:
{{range .PeriodTotals}}  {{.Category.Name}}: {{duration .Minutes}}
{{end}}
Daily breakdown:
{{range .DailyTotals}}{{.Date}} — {{duration .ActualMinutes}} / {{duration .TargetMinutes}} target
{{if .Categories}}{{range .Categories}}  {{.Category.Name}}: {{duration .Minutes}}
{{end}}{{else}}  (no tracked time)
{{end}}
{{end}}'
WHERE key = 'text_summary' AND builtin = 1;

package google

// Google Calendar API response shapes for the endpoints this provider uses.

type calendarListResponse struct {
	Items         []calendarListEntry `json:"items"`
	NextPageToken string              `json:"nextPageToken"`
}

type calendarListEntry struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Primary bool   `json:"primary"`
}

type eventsListResponse struct {
	Items         []event `json:"items"`
	NextPageToken string  `json:"nextPageToken"`
}

type event struct {
	ID                string     `json:"id"`
	RecurringEventID  string     `json:"recurringEventId"`
	ICalUID           string     `json:"iCalUID"`
	Summary           string     `json:"summary"`
	Description       string     `json:"description"`
	Location          string     `json:"location"`
	Status            string     `json:"status"`
	Start             eventTime  `json:"start"`
	End               eventTime  `json:"end"`
	OriginalStartTime *eventTime `json:"originalStartTime"`
	Organizer         *person    `json:"organizer"`
	Attendees         []attendee `json:"attendees"`
}

type eventTime struct {
	Date     string `json:"date"`     // all-day: YYYY-MM-DD
	DateTime string `json:"dateTime"` // timed: RFC3339
	TimeZone string `json:"timeZone"`
}

type person struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type attendee struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	ResponseStatus string `json:"responseStatus"`
	Organizer      bool   `json:"organizer"`
	Self           bool   `json:"self"`
}

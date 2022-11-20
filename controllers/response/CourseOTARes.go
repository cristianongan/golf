package response

type CourseOTARes struct {
	Code        string `json:"Code"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Image       string `json:"Image"`
	Logo        string `json:"Logo"`
	Holes       int    `json:"Holes"`
}

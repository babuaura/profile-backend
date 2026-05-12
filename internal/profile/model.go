package profile

type Profile struct { Name string `json:"name"`; Title string `json:"title"`; Location string `json:"location"`; Email string `json:"email"`; Website string `json:"website"`; Summary string `json:"summary"`; Highlights []Metric `json:"highlights"`; Links []ProfileLink `json:"links"`; FocusAreas []string `json:"focusAreas"` }
type Metric struct { Label string `json:"label"`; Value string `json:"value"` }
type ProfileLink struct { Label string `json:"label"`; URL string `json:"url"` }

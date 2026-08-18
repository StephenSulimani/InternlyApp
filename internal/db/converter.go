package db

func ToCreateParams(j Job) CreateJobParams {
	return CreateJobParams{
		SourceUrl:       j.SourceUrl,
		SourceName:      j.SourceName,
		FirstSeen:       j.FirstSeen,
		ApplicationLink: j.ApplicationLink,
		Company:         j.Company,
		RoleTitle:       j.RoleTitle,
		Locations:       j.Locations,
		JobType:         j.JobType,
		IsAts:           j.IsAts,
		Column10:        j.Metadata,
	}
}

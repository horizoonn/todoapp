package users

type GetUsersFilter struct {
	Limit  int
	Offset int
}

func NewGetUsersFilter(limit int, offset int) GetUsersFilter {
	return GetUsersFilter{
		Limit:  limit,
		Offset: offset,
	}
}

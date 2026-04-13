package types

type LoginReq struct {
	Email string `json:"email" validate:"required,email"`
}

type CreateUserReq struct {
	Name string `json:"name" validate:"required,min=2"`
}

type ProfileResp struct {
	Name string `json:"name"`
}

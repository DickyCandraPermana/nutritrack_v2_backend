package domain

type UserCreateInput struct {
	Username string `validate:"required,min=3,max=30"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

type UserUpdateInput struct {
	Username *string `validate:"min=3,max=30"`
	Email    *string `validate:"email"`
}

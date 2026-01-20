package auth

func (h *Handler) Register(email string, username string, password string) bool{
	if email == "" || username == "" || password == "" {
		return false;
	}
}
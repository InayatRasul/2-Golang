package modules
import(
	"time"
)

type User struct {
	ID int `db:"id"`
	Name string `db:"name"`
	Email     string    `db:"email" json:"email"`
	Age       int       `db:"age" json:"age"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

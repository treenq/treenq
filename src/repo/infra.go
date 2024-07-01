package repo

import (
	"database/sql"
	"fmt"
)


func Update(cr *sql.DB,table string,condition string,column string,updatedValue string){
	query := fmt.Sprintf("",)
	cr.Exec(query,table,condition,column,updatedValue)
}  
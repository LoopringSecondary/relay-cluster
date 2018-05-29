dialects=("postgres" "mysql" "mssql" "sqlite")

for dialect in "${dialects[@]}" ; do
<<<<<<< HEAD
    DEBUG=false GORM_DIALECT=${dialect} go test
=======
    GORM_DIALECT=${dialect} go test
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
done

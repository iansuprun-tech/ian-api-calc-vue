// Пакет db экспортирует встроенные SQL-файлы миграций.
// go:embed работает только с путями относительно файла, в котором объявлен,
// поэтому embed.FS живёт здесь — рядом с папкой migrations/.
package db

import "embed"

//go:embed migrations/*.sql
var MigrationsFS embed.FS

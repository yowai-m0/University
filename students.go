package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
    
    "github.com/joho/godotenv"
	"github.com/jackc/pgx/v5"
)

func main() {
    reader := bufio.NewReader(os.Stdin)
    // ===== 1. ЗАГРУЖАЕМ .env ФАЙЛ =====
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Ошибка загрузки .env файла:", err)
    }
    
    // ===== 2. ЧИТАЕМ ПЕРЕМЕННЫЕ =====
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    dbname := os.Getenv("DB_NAME")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    
    // ===== 3. СОБИРАЕМ СТРОКУ ПОДКЛЮЧЕНИЯ =====
    connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
        host, port, dbname, user, password)
    
    // ===== 4. ПОДКЛЮЧАЕМСЯ =====
    conn, err := pgx.Connect(context.Background(), connStr)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Успешное подключение к БД") 

    fmt.Println()

    fmt.Println("Создаем БД")
    _, err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS students (
            id SERIAL PRIMARY KEY,
            name TEXT,
            age INT,
            course INT
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    _, err = conn.Exec(context.Background(), "DELETE FROM students")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Таблица очищена")

    _, err = conn.Exec(context.Background(),
        "ALTER SEQUENCE students_id_seq RESTART WITH 1")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Счетчик обновлен успешно")

    students := []struct {
        name string
        age int
        course int
    }{
        {"Иван Петров", 19, 2},
		{"Анна Смирнова", 20, 3},
		{"Петр Иванов", 21, 4},
		{"Алексей Иванов", 19, 2},
		{"Мария Смирнова", 18, 1},
		{"Дмитрий Кузнецов", 22, 4},
		{"Анна Попова", 17, 1},
		{"Максим Васильев", 20, 3},
		{"Екатерина Соколова", 23, 5},
		{"Артем Михайлов", 21, 3},
		{"Виктория Новикова", 19, 2},
		{"Илья Федоров", 24, 6},
		{"Полина Морозова", 18, 1},
		{"Никита Волков", 20, 2},
		{"Анастасия Алексеева", 22, 4},
		{"Кирилл Лебедев", 17, 1},
		{"Дарья Семенова", 21, 3},
		{"Егор Павлов", 25, 6},
		{"София Козлова", 19, 2},
		{"Владислав Степанов", 22, 4},
		{"Алина Николаева", 18, 1},
		{"Михаил Орлов", 23, 5},
		{"Полина Андреева", 20, 3},
		{"Александр Макаров", 17, 1},
		{"Валерия Захарова", 21, 3},
		{"Роман Чернышёв", 24, 5},
		{"Елизавета Калинина", 19, 2},
		{"Игорь Фролов", 25, 6},
    }

    for _, s := range students {
        _, err = conn.Exec(context.Background(),
            "INSERT INTO students (name, age, course) VALUES ($1, $2, $3)",
            s.name, s.age, s.course)
        if err != nil {
            log.Fatal(err)
        }
    }
    fmt.Println("Успешно добавленые студенты")

    var count int
    err = conn.QueryRow(context.Background(), 
        "SELECT COUNT(*) FROM students").Scan(&count)
    if err != nil {
        log.Fatal(err)
    } 
    fmt.Printf("Всего записей: %d\n", count)  

    for {
        Menu()
        fmt.Println("Выберите действие")
        var choise int
        fmt.Scan(&choise)
        fmt.Scanln()

        switch choise {
        case 1:
            row, err := conn.Query(context.Background(),
                "SELECT id, name, age, course FROM students ORDER BY id")
            if err != nil {
                log.Fatal(err)
            }
            defer row.Close()

            fmt.Println("=== Список студентов ===")
            fmt.Println(" ID  |  Name  |  Age  |  Course")

            for row.Next() {
                var id int
                var name string
                var age int
                var course int

                err = row.Scan(&id, &name, &age, &course)
                if err != nil {
                    log.Fatal(err)
                }
                fmt.Printf(" %d  |  %s  |  %d  |  %d\n", id, name, age, course)
            }
            fmt.Println()

        case 2:
            fmt.Println("Введите имя и фамилию студента")
            var nameStud string
            nameStud, _ = reader.ReadString('\n')
            nameStud = strings.TrimSpace(nameStud)

            fmt.Println("Введите возраст")
            var ageStud int
            fmt.Scan(&ageStud)

            fmt.Println("Введите курс обучения")
            var courseStud int
            fmt.Scan(&courseStud)

            _, err = conn.Exec(context.Background(),
                "INSERT INTO students (name, age, course) VALUES ($1, $2, $3)",
                nameStud, ageStud, courseStud)
            if err != nil {
                log.Fatal(err)
            }      
            fmt.Printf("Студент %s успешно добавлен\n", nameStud)
            fmt.Println()

        case 3:
            fmt.Println("Введите курс чтобы посмотреть студентов")
            var courseFind int
            fmt.Scan(&courseFind)

            rows, err := conn.Query(context.Background(),
                "SELECT name, age FROM students WHERE course = $1", courseFind)
            if err != nil {
                log.Fatal(err)
            }   

            fmt.Printf("Студенты %d курса:\n", courseFind)
            found := false
            for rows.Next() {
                var name string
                var age int
                err = rows.Scan(&name, &age)
                if err != nil {
                    log.Fatal(err)
                }
                fmt.Printf(" - %s (%d лет)\n", name, age)
                found = true
            }

            if !found {
                fmt.Printf("Студенты на этом курсе(%d) не найдены\n", courseFind)
            }
            fmt.Println()

        case 4:
            fmt.Println("Введите ID студента, чтобы обновить его курс")
            var idStud int
            fmt.Scan(&idStud)

            fmt.Println("Введите нужный курс(1-6)")
            var courseNew int
            fmt.Scan(&courseNew)

            res, err := conn.Exec(context.Background(),
                "UPDATE students SET course = $1 WHERE id = $2", courseNew, idStud)
            if err != nil {
                log.Fatal(err)
            }

            if res.RowsAffected() > 0 {
                fmt.Println("Курс обновлен")
            }else {
                fmt.Println("Студент не найден")
            }
            fmt.Println()

        case 5:
            fmt.Println("Введите ID студента для удаления")
            var nowId int
            fmt.Scan(&nowId)

            res, err := conn.Exec(context.Background(),
                "DELETE FROM students WHERE id = $1", nowId)
            if err != nil {
                log.Fatal(err)
            }

            if res.RowsAffected() > 0 {
                fmt.Println("Студент успешно удален")
            }else {
                fmt.Println("Студент не найден")
            }
        case 0:
            fmt.Println("ПОКА!")
            conn.Close(context.Background())
            fmt.Println("Таблица закрыта успешно")
            return
        }
    }
}

func Menu() {
    fmt.Println("====== Menu ======")
    fmt.Println("1. Показать всех студентов")
    fmt.Println("2. Добавить студента")
    fmt.Println("3. Найти по курсу")
    fmt.Println("4. Обновить курс")
    fmt.Println("5. Удалить студента")
    fmt.Println("0. Exit")
}
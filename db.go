package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Dish struct {
	ID          int64
	Name        string
	Description string
	Price       string // только число: "690", "1 890", "180 / 250"
	Image       string
	Badge       string // "Хит", "Новинка", пусто
}

type Section struct {
	ID     int64
	Slug   string
	Title  string
	Number string // "01", "02", ...
	Items  []Dish
}

func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	if err := seed(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sections (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			slug       TEXT UNIQUE NOT NULL,
			title      TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS dishes (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			section_id  INTEGER NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
			name        TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			price       TEXT NOT NULL DEFAULT '',
			image       TEXT NOT NULL DEFAULT '',
			badge       TEXT NOT NULL DEFAULT '',
			sort_order  INTEGER NOT NULL DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_dishes_section ON dishes(section_id);
	`)
	return err
}

func seed(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sections`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	type seedDish struct {
		Name, Description, Price, Image, Badge string
	}
	type seedSection struct {
		Slug, Title string
		Dishes      []seedDish
	}

	data := []seedSection{
		{"starters", "Закуски", []seedDish{
			{"Тартар из говядины", "Вырезка, каперсы, дижонский майонез, тосты из бородинского хлеба.", "690", "/img/Тартар из говядины.jpg", "Хит"},
			{"Сырная тарелка", "Пять сортов сыра, мёд с грецким орехом, виноград, домашнее варенье.", "890", "/img/Сырная тарелка.png", ""},
			{"Брускетта с томатами", "Хрустящий хлеб, черри, базилик, чеснок, оливковое масло.", "420", "/img/Брускетта с томатами.png", ""},
		}},
		{"salads", "Салаты", []seedDish{
			{"Цезарь с курицей", "Романо, куриная грудка на гриле, пармиджано, крутоны, соус цезарь.", "590", "/img/Цезарь с курицей.png", ""},
			{"Греческий", "Огурец, томаты, фета, маслины, красный лук, орегано, оливковое масло.", "480", "/img/Греческий салат.png", ""},
			{"Хрустящие баклажаны", "Баклажаны в панировке, томаты, кинза, йогуртовый соус с чесноком.", "620", "/img/Баклажаны.png", "Новинка"},
		}},
		{"soups", "Супы", []seedDish{
			{"Крем-суп из тыквы", "Запечённая тыква, сливки, тыквенные семечки, трюфельное масло.", "380", "/img/Крем-суп из тыквы.png", ""},
			{"Том-ям с креветками", "Кокосовое молоко, тигровые креветки, лемонграсс, шампиньоны, чили.", "690", "/img/Том-ям с креветками.png", ""},
			{"Борщ с говядиной", "Наваристый бульон, свежая свёкла, сметана, пампушка с чесноком.", "450", "/img/Борщ с говядиной.png", ""},
			{"Куриный бульон с лапшой", "Домашняя лапша, курица, морковь, зелень, перепелиное яйцо.", "340", "", ""},
		}},
		{"hot", "Горячее", []seedDish{
			{"Стейк рибай", "Мраморная говядина, соус демиглас, печёный картофель, розмарин.", "1 890", "", "Хит"},
			{"Лосось на гриле", "Стейк лосося, спаржа, лимонно-сливочный соус, укроп.", "1 240", "", ""},
			{"Паста карбонара", "Спагетти, гуанчале, желток, пекорино романо, чёрный перец.", "690", "", ""},
			{"Утиная грудка", "Утиное филе, вишнёвый соус, пюре из батата, свежий тимьян.", "980", "", ""},
		}},
		{"desserts", "Десерты", []seedDish{
			{"Чизкейк Нью-Йорк", "Классический сливочный чизкейк, песочная основа, ягодный соус.", "420", "", ""},
			{"Тирамису", "Маскарпоне, савоярди, эспрессо, какао.", "450", "", ""},
			{"Шоколадный фондан", "Тёплый шоколадный бисквит с жидким центром, шарик мороженого.", "490", "", "Хит"},
			{"Медовик", "Домашние медовые коржи, сметанный крем, грецкий орех.", "380", "", ""},
		}},
		{"drinks", "Напитки", []seedDish{
			{"Эспрессо / Капучино", "Свежеобжаренное зерно, альтернативные варианты на овсяном молоке.", "180 / 250", "", ""},
			{"Домашний лимонад", "На выбор: облепиха–апельсин, лаванда–лимон, малина–мята.", "320", "", ""},
			{"Свежевыжатый сок", "Апельсин, грейпфрут, морковь, яблоко или сельдерей.", "380", "", ""},
			{"Бокал вина", "Красное или белое сухое. Спросите официанта о винной карте.", "450", "", ""},
		}},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for si, s := range data {
		res, err := tx.Exec(`INSERT INTO sections(slug, title, sort_order) VALUES(?,?,?)`, s.Slug, s.Title, si)
		if err != nil {
			return err
		}
		sectionID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		for di, d := range s.Dishes {
			if _, err := tx.Exec(`
				INSERT INTO dishes(section_id, name, description, price, image, badge, sort_order)
				VALUES(?,?,?,?,?,?,?)`,
				sectionID, d.Name, d.Description, d.Price, d.Image, d.Badge, di); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func LoadMenu(db *sql.DB) ([]Section, error) {
	sectionRows, err := db.Query(`SELECT id, slug, title FROM sections ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer sectionRows.Close()

	var sections []Section
	index := map[int64]int{}
	for sectionRows.Next() {
		var s Section
		if err := sectionRows.Scan(&s.ID, &s.Slug, &s.Title); err != nil {
			return nil, err
		}
		sections = append(sections, s)
		index[s.ID] = len(sections) - 1
	}
	if err := sectionRows.Err(); err != nil {
		return nil, err
	}

	for i := range sections {
		sections[i].Number = fmt.Sprintf("%02d", i+1)
	}

	dishRows, err := db.Query(`
		SELECT section_id, name, description, price, image, badge
		FROM dishes
		ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer dishRows.Close()

	for dishRows.Next() {
		var sectionID int64
		var d Dish
		if err := dishRows.Scan(&sectionID, &d.Name, &d.Description, &d.Price, &d.Image, &d.Badge); err != nil {
			return nil, err
		}
		if i, ok := index[sectionID]; ok {
			sections[i].Items = append(sections[i].Items, d)
		}
	}
	return sections, dishRows.Err()
}

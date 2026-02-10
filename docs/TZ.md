# Технічне завдання: MVP системи паспортного столу

## 1. Загальна інформація

**Назва проєкту:** Passport Desk MVP  
**Тип застосунку:** Десктопний застосунок для Windows/Linux  
**Цільова аудиторія:** Один оператор паспортного столу  
**Мета:** Облік громадян, реєстрація за місцем проживання/перебування, формування звітності

---

## 2. Технологічний стек

### Backend
- **Мова:** Go 1.21+
- **Framework:** Wails v2
- **База даних:** SQLCipher (SQLite з шифруванням AES-256)
- **Шифрування:** Go crypto/aes, crypto/cipher, golang.org/x/crypto/pbkdf2

### Frontend
- **Framework:** Vue.js 3 
- **UI Kit:** Naive UI 
- **Стилі:** Tailwind CSS

### Бібліотеки
```
github.com/wailsapp/wails/v2
github.com/mutecomm/go-sqlcipher/v4
github.com/zalando/go-keyring (для зберігання ключів)
github.com/jung-kurt/gofpdf (генерація PDF-звітів)
```

---

## 3. Функціональні вимоги

### 3.1 Автентифікація та безпека

**FR-1.1** Вхід в систему
- Логін та пароль оператора
- Пароль використовується для генерації ключа шифрування БД (PBKDF2, 100000 ітерацій)
- Зберігання хешу пароля в окремій незашифрованій таблиці (bcrypt)

**FR-1.2** Автоблокування
- Блокування застосунку після 5 хвилин неактивності
- Повторний запит пароля для розблокування

**FR-1.3** Зміна пароля
- Можливість зміни пароля оператора
- Автоматичне перешифрування БД з новим ключем

### 3.2 Облік громадян

**FR-2.1** Додавання громадянина
Обов'язкові поля:
- Прізвище, ім'я, по батькові
- Дата народження
- Серія та номер паспорта (зашифровано на рівні застосунку)
- ІПН/РНОКПП (зашифровано на рівні застосунку)
- Стать

Додаткові поля:
- Місце народження
- Контактний телефон (зашифровано)
- Email
- Примітки

**FR-2.2** Реєстрація за адресою
- Тип реєстрації: постійна/тимчасова
- Адреса (повна): область, район, населений пункт, вулиця, будинок, квартира
- Дата реєстрації
- Дата зняття з реєстрації (для тимчасової)
- Підстава реєстрації (документ)

**FR-2.3** Пошук громадян
Пошук за:
- ПІБ (частковий збіг)
- Паспортом
- ІПН
- Адресою
- Датою народження

**FR-2.4** Редагування даних
- Редагування всіх полів громадянина
- Історія змін (аудит-лог)

**FR-2.5** Видалення/архівування
- М'яке видалення (позначка deleted = true)
- Можливість відновлення

### 3.3 Звітність

**FR-3.1** Список зареєстрованих громадян
- Фільтрація за періодом
- Експорт у PDF/Excel

**FR-3.2** Довідка про реєстрацію
- Генерація PDF-довідки з даними громадянина
- Номер довідки, дата видачі, печатка (зображення)

**FR-3.3** Статистика
- Кількість зареєстрованих за період
- Розподіл за типом реєстрації
- Загальна кількість активних реєстрацій

### 3.4 Аудит та логування

**FR-4.1** Журнал операцій
Запис всіх операцій:
- Дата та час
- Оператор (ім'я користувача)
- Тип операції (CREATE, READ, UPDATE, DELETE)
- ID запису
- Опис дії

**FR-4.2** Перегляд логів
- Фільтрація за датою, типом операції
- Експорт у файл

---

## 4. Структура бази даних

### Таблиця: `operators`
```sql
CREATE TABLE operators (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,  -- bcrypt
    full_name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Таблиця: `citizens`
```sql
CREATE TABLE citizens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    last_name TEXT NOT NULL,
    first_name TEXT NOT NULL,
    middle_name TEXT,
    birth_date DATE NOT NULL,
    passport_series TEXT NOT NULL,      -- зашифровано
    passport_number TEXT NOT NULL,      -- зашифровано
    tax_number TEXT,                    -- ІПН, зашифровано
    gender TEXT CHECK(gender IN ('M', 'F')),
    birth_place TEXT,
    phone TEXT,                         -- зашифровано
    email TEXT,
    notes TEXT,
    deleted BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Таблиця: `registrations`
```sql
CREATE TABLE registrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    citizen_id INTEGER NOT NULL,
    registration_type TEXT CHECK(registration_type IN ('permanent', 'temporary')),
    region TEXT NOT NULL,               -- область
    district TEXT,                      -- район
    settlement TEXT NOT NULL,           -- населений пункт
    street TEXT NOT NULL,
    house_number TEXT NOT NULL,
    apartment_number TEXT,
    registration_date DATE NOT NULL,
    deregistration_date DATE,
    basis_document TEXT,                -- підстава
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (citizen_id) REFERENCES citizens(id)
);
```

### Таблиця: `audit_log`
```sql
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    operator_id INTEGER NOT NULL,
    action_type TEXT NOT NULL,          -- CREATE, READ, UPDATE, DELETE
    table_name TEXT NOT NULL,
    record_id INTEGER,
    description TEXT,
    FOREIGN KEY (operator_id) REFERENCES operators(id)
);
```

### Таблиця: `certificates`
```sql
CREATE TABLE certificates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    certificate_number TEXT UNIQUE NOT NULL,
    citizen_id INTEGER NOT NULL,
    issue_date DATE NOT NULL,
    purpose TEXT,
    issued_by INTEGER NOT NULL,         -- operator_id
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (citizen_id) REFERENCES citizens(id),
    FOREIGN KEY (issued_by) REFERENCES operators(id)
);
```

---

## 5. Архітектура безпеки

### 5.1 Рівні шифрування

**Рівень 1: Шифрування файлу БД (SQLCipher)**
- Весь файл `passport_desk.db` зашифрований AES-256
- Ключ генерується з пароля оператора: `PBKDF2(password, salt, 100000, SHA-256)`
- Salt зберігається в окремому файлі `db.salt`

**Рівень 2: Шифрування полів (Application layer)**
Зашифровані поля:
- `passport_series`, `passport_number`
- `tax_number`
- `phone`

Алгоритм: AES-256-GCM  
Ключ: той самий, що й для SQLCipher (кешується в пам'яті)

### 5.2 Управління ключами

```go
// Псевдокод
type SecurityManager struct {
    masterKey []byte  // Зберігається тільки в RAM
    salt      []byte  // Завантажується з db.salt
}

func (sm *SecurityManager) Login(password string) error {
    // 1. Генерація ключа з пароля
    sm.masterKey = pbkdf2.Key([]byte(password), sm.salt, 100000, 32, sha256.New)
    
    // 2. Відкриття БД
    db, err := sql.Open("sqlite3", "file:passport_desk.db?_pragma_key=x'"+hex.EncodeToString(sm.masterKey)+"'")
    
    // 3. Перевірка пароля (спроба SELECT)
    return err
}

func (sm *SecurityManager) EncryptField(plaintext string) (string, error) {
    // AES-256-GCM з masterKey
}
```

### 5.3 Зберігання ключів

- **Windows:** Використання DPAPI через `github.com/zalando/go-keyring`
- **Linux:** Використання Secret Service API (GNOME Keyring)
- **Fallback:** Зашифрований файл з ключем (ризиковано, але для MVP прийнятно)

---

## 6. Користувацький інтерфейс

### 6.1 Екран входу
- Поля: логін, пароль
- Кнопка "Увійти"
- Повідомлення про помилки

### 6.2 Головне вікно
```
┌─────────────────────────────────────────────────┐
│ [Паспортний стіл]              [Оператор: Іван] │
├─────────────────────────────────────────────────┤
│ [Громадяни] [Реєстрації] [Звіти] [Логи]        │
├─────────────────────────────────────────────────┤
│                                                  │
│  [Пошук: _______________] [Знайти] [+ Новий]   │
│                                                  │
│  ┌────────────────────────────────────────────┐ │
│  │ ПІБ          │ Паспорт  │ Дата народж. │→│ │
│  ├────────────────────────────────────────────┤ │
│  │ Іваненко ... │ СР123456 │ 01.01.1990   │ │ │
│  │ Петренко ... │ НК654321 │ 15.05.1985   │ │ │
│  └────────────────────────────────────────────┘ │
│                                                  │
│                                    [Вихід]       │
└─────────────────────────────────────────────────┘
```

### 6.3 Форма додавання/редагування громадянина
- Вкладки: "Особисті дані", "Реєстрація", "Історія"
- Валідація полів (формат паспорта, ІПН, дати)
- Кнопки: "Зберегти", "Скасувати"

### 6.4 Модуль звітів
- Вибір типу звіту
- Параметри фільтрації (дати, тип реєстрації)
- Попередній перегляд
- Експорт (PDF, XLSX)

---

## 7. Нефункціональні вимоги

### 7.1 Продуктивність
- Пошук по 10,000 записах: < 500 мс
- Відкриття форми редагування: < 200 мс
- Генерація звіту (100 записів): < 2 с

### 7.2 Безпека
- Весь трафік між frontend та backend всередині процесу (Wails)
- Зашифровані поля не логуються у відкритому вигляді
- Автоматичне очищення буферу обміну з чутливими даними (через 30 с)

### 7.3 Надійність
- Автоматичний backup БД щоденно
- Backup зберігається в папці `backups/` з датою у назві
- Можливість відновлення з backup

### 7.4 Зручність
- Інтуїтивний інтерфейс
- Гарячі клавіші (Ctrl+N - новий, Ctrl+F - пошук, Ctrl+S - зберегти)
- Підказки (tooltips) для всіх полів

---

## 8. Етапи розробки MVP

### Етап 1: Основа (тиждень 1)
- [ ] Налаштування проєкту Wails
- [ ] Підключення SQLCipher
- [ ] Реалізація автентифікації
- [ ] Базовий UI (логін, головне вікно)

### Етап 2: CRUD громадян (тиждень 2)
- [ ] Таблиця `citizens`
- [ ] Форма додавання/редагування
- [ ] Шифрування чутливих полів
- [ ] Пошук

### Етап 3: Реєстрації (тиждень 3)
- [ ] Таблиця `registrations`
- [ ] Додавання реєстрації до громадянина
- [ ] Перегляд історії реєстрацій
- [ ] Зняття з реєстрації

### Етап 4: Звітність (тиждень 4)
- [ ] Генерація PDF-довідок
- [ ] Список зареєстрованих
- [ ] Статистика
- [ ] Експорт у Excel

### Етап 5: Аудит та фінальні доопрацювання (тиждень 5)
- [ ] Аудит-лог всіх операцій
- [ ] Автоматичний backup
- [ ] Тестування безпеки
- [ ] Документація користувача

---

## 9. Структура проєкту

```
passport-desk-mvp/
├── main.go
├── wails.json
├── app.go                    # Wails app structure
├── frontend/
│   ├── src/
│   │   ├── App.vue
│   │   ├── views/
│   │   │   ├── Login.vue
│   │   │   ├── Dashboard.vue
│   │   │   ├── CitizenList.vue
│   │   │   ├── CitizenForm.vue
│   │   │   ├── Reports.vue
│   │   │   └── AuditLog.vue
│   │   └── components/
│   └── package.json
├── internal/
│   ├── database/
│   │   ├── db.go
│   │   ├── migrations.go
│   │   └── models.go
│   ├── security/
│   │   ├── crypto.go
│   │   ├── auth.go
│   │   └── keystore.go
│   ├── services/
│   │   ├── citizen_service.go
│   │   ├── registration_service.go
│   │   ├── report_service.go
│   │   └── audit_service.go
│   └── utils/
│       ├── validator.go
│       └── backup.go
├── backups/
├── docs/
│   └── user_manual.md
└── README.md
```

---

## 10. Критерії прийняття MVP

✅ Оператор може увійти в систему  
✅ Додавання нового громадянина зі всіма обов'язковими полями  
✅ Пошук громадян за ПІБ, паспортом, ІПН  
✅ Редагування даних громадянина  
✅ Додавання постійної/тимчасової реєстрації  
✅ Генерація PDF-довідки про реєстрацію  
✅ Перегляд списку зареєстрованих за період  
✅ Всі чутливі дані зашифровані (паспорт, ІПН, телефон)  
✅ БД повністю зашифрована SQLCipher  
✅ Автоблокування після 5 хв неактивності  
✅ Аудит-лог всіх операцій  
✅ Автоматичний щоденний backup  
✅ Застосунок працює на Windows 10/11 та Ubuntu 22.04+  

---

## 11. Ризики та обмеження

### Ризики
1. **Втрата пароля** → втрата доступу до всіх даних (рішення: процедура відновлення з майстер-ключем)
2. **Пошкодження БД** → втрата даних (рішення: автоматичні backups)
3. **Недостатня продуктивність** SQLCipher на великих обсягах (рішення: оптимізація запитів, індекси)

### Обмеження MVP
- Тільки один оператор (без багатокористувацького режиму)
- Без інтеграції з державними реєстрами
- Без друку довідок (тільки PDF)
- Базова статистика (без складних аналітичних звітів)

---

## 12. Подальший розвиток (post-MVP)

- Багатокористувацький режим (PostgreSQL замість SQLite)
- Інтеграція з ДРФО, ЄДДР
- Розширена аналітика
- Мобільний додаток для операторів
- Веб-версія для керівництва
- Електронний документообіг

---

## 13. Контакти та питання

При виникненні питань або необхідності уточнень звертайтеся до розробника.

**Дата створення ТЗ:** 01.02.2026  
**Версія:** 1.0  
**Статус:** Затверджено до розробки
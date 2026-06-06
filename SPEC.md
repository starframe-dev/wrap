# Warp — Layout Engine для терминальных TUI

## Контекст

Automata — AI-агент на базе pi.dev — нуждается в терминальном интерфейсе с гибким управлением
пространством: вкладки, разделение панелей, плавающие окна. Существующий `viewer` из tui-launcher
жёстко завязан на эмуляцию терминала (VT10x + PTY).

Нужен **чистый layout-движок** — библиотека, которая управляет только расположением панелей,
оставляя их содержимое на усмотрение пользователя.

## Цель

Создать Go-модуль `github.com/Starframe/warp` — Bubbletea-based layout engine с поддержкой
tabs, splits (vertical/horizontal), drag-and-drop ресайза границ и плавающих панелей (float panes).

## Модуль

- **Путь:** `github.com/Starframe/warp`
- **Директория:** `/Users/a/Space/Projects/Starframe/warp`
- **Язык:** Go 1.22+
- **Зависимости:** `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`

## API — Концепт

```go
package warp

// Panel — интерфейс, который реализует пользователь для своих панелей.
// Панель может быть чем угодно: терминал, текст, графика, форма.
type Panel interface {
    // View рендерит содержимое панели заданного размера.
    View(width, height int) string

    // Update обрабатывает сообщения Bubbletea (клавиши, мышь, etc).
    // Панель получает только те сообщения, которые пришли, когда она была в фокусе.
    Update(msg tea.Msg) tea.Cmd
}

// Warp — корневая структура layout-движка. Реализует tea.Model.
type Warp struct { ... }

func New() *Warp

// Tab — вкладка, содержит layout из панелей.
type Tab struct { ... }

func (w *Warp) NewTab(name string) *Tab

// Layout-операции внутри Tab:

// SplitVertical разделяет панель по вертикали (лево/право).
// fraction — доля для левой панели (0.0–1.0).
func (t *Tab) SplitVertical(parent Panel, fraction float64, newPanel Panel)

// SplitHorizontal разделяет панель по горизонтали (верх/низ).
// fraction — доля для верхней панели (0.0–1.0).
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, newPanel Panel)

// Float делает панель плавающей (поверх остальных).
func (t *Tab) Float(panel Panel, x, y, width, height int)

// TabPosition — расположение таб-бара.
type TabPosition int

const (
    TabTop    TabPosition = iota // сверху (по умолчанию)
    TabBottom                    // снизу
    TabLeft                      // слева
    TabRight                     // справа
)

// Запуск:
func (w *Warp) Run() error

// Настройка:
func (w *Warp) SetTabPosition(pos TabPosition)
```

## Что изменится

1. `warp.go` — корневая структура `Warp`, `tea.Model`, управление табами
2. `tab.go` — структура `Tab`, дерево панелей, splits
3. `panel.go` — интерфейс `Panel`, базовая реализация `BasePanel`
4. `split.go` — структура `Split` (контейнер с direction + children)
5. `float.go` — логика плавающих панелей (z-order, перемещение, ресайз)
6. `drag.go` — drag-and-drop ресайз границ между панелями
7. `render.go` — рендеринг дерева панелей в строки
8. `styles.go` — стили для таб-бара, границ, float-рамок
9. `go.mod` — модуль

## Детали реализации

### 1. Дерево панелей (Tree of Panels)

Внутри каждого `Tab` панели организованы в бинарное дерево splits:

```go
type Node struct {
    // Terminal node — содержит Panel
    Panel Panel

    // Internal node — содержит split
    Split *SplitConfig
}

type SplitConfig struct {
    Direction  Direction  // Vertical или Horizontal
    Fraction   float64     // Доля первого child (0.0–1.0)
    First      *Node
    Second     *Node
    Dragging   bool        // true во время drag-and-drop
}
```

Корень дерева — всегда `SplitConfig` (даже если одна панель: SplitConfig с одним child = Panel).

### 2. Рендеринг (Layout → строки)

Рекурсивный алгоритм:

```
render(node, x, y, w, h):
    если node.Panel != nil:
        вернуть node.Panel.View(w, h)
    если node.Split != nil:
        firstW, secondW, firstH, secondH = вычислить(node.Split, w, h)
        first := render(node.Split.First, x, y, firstW, firstH)
        second := render(node.Split.Second, x2, y2, secondW, secondH)
        если Vertical:  склеить горизонтально (построчно)
        если Horizontal: склеить вертикально
```

Границы между панелями — 1 символ: `│` для вертикальных, `─` для горизонтальных.

### 3. Drag-and-drop границ

- Мышь над границей → курсор меняется (визуально, через стиль границы)
- MousePress на границе → начало drag
- MouseMotion → изменение Fraction у SplitConfig, перерендер
- MouseRelease → конец drag
- Минимальный размер панели: 3 символа (чтобы не схлопывалась в ноль)

### 4. Плавающие панели (Float panes)

- Хранятся отдельным списком в `Tab` с z-order (последняя созданная — сверху)
- У float-панели: позиция (x, y), размер (w, h), заголовок
- Заголовок float-панели можно drag-and-drop для перемещения
- Границы float-панели можно drag-and-drop для ресайза
- Float-панели рендерятся поверх основного layout после его отрисовки
- Перекрытые float-панелями области основного layout затираются

### 5. Фокус и навигация

- `Ctrl+Tab` / `Ctrl+Shift+Tab` — переключение табов
- `Ctrl+W` — закрыть текущий таб
- `Ctrl+T` — новый таб
- Клик мышью по панели — фокус на неё
- Сообщения `tea.Msg` пробрасываются только сфокусированной панели
- Для float-панелей: клик по float-панели = фокус на неё. Она остаётся сверху.

### 6. Таб-бар

- Настраиваемая позиция: сверху, снизу, слева, справа (по умолчанию сверху)
- `TabTop`/`TabBottom`: горизонтальный таб-бар, высота 1 строка
- `TabLeft`/`TabRight`: вертикальный таб-бар, ширина вычисляется по длине имён табов
- Формат: `[tab1] [tab2] [active] [+]`
- `+` — кнопка нового таба
- `×` — кнопка закрытия таба (на активном табе)
- Клик по табу переключает на него

### 7. Выпадающие меню (Dropdown menus)

- Меню открывается из кнопки в таб-баре (левее `+`)
- Список пунктов: имя + горячая клавиша
- Курсор навигации (↑↓), Enter — выбрать, Esc — закрыть
- Отображается под кнопкой (или над, если нет места)

Модель:
```go
type MenuItem struct {
    Name    string
    Key     string    // Горячая клавиша (опционально)
    Action  func() tea.Cmd
    Enabled bool
}

type Menu struct {
    Items   []MenuItem
    X, Y    int       // Позиция меню
    Width   int
    Cursor  int
    Open    bool
}
```

### 8. Контекстные меню (Context menus)

- По правому клику (MouseButtonRight) на панели
- Появляется в точке клика
- Те же MenuItem, что и в dropdown
- Закрывается по Esc или левому клику вне меню

### 9. Скролл панелей

- Если Panel.View возвращает больше строк, чем выделено места — панель скроллится
- PageUp/PageDown/колесо мыши для скролла
- Визуальный скроллбар (правый край панели)
- Скролл хранится на уровне Node (каждый лист = свой scroll offset)

```go
type ScrollableNode struct {
    Node
    ScrollOffset int
    TotalLines   int  // Общее число строк, возвращённых Panel.View
}
```

### 10. Word wrap (перенос слов)

- По умолчанию Panel.View(w, h) возвращает h строк шириной w
- Если контент не помещается по ширине — warp может сделать word wrap
- Алгоритм: разбить строку на слова, укладывать в w, переносить не влезающее
- Опционально: пользователь может отключить wrap для панели

## Критерии приёмки

- [x] `warp.New()` создаёт экземпляр с одним пустым табом
- [x] `NewTab(name)` добавляет вкладку, переключает на неё
- [x] `SplitVertical(panel, 0.5, newPanel)` — две панели слева/справа, граница `│`
- [x] `SplitHorizontal(panel, 0.5, newPanel)` — две панели сверху/снизу, граница `─`
- [x] Drag-and-drop границы мышью меняет пропорции панелей
- [x] Минимальный размер панели — 3 символа
- [x] `Float(panel, x, y, w, h)` — панель появляется поверх layout
- [x] Float-панель можно двигать за заголовок мышью
- [x] Float-панель можно ресайзить за границы мышью
- [x] Таб-бар показывает все вкладки, клик переключает
- [x] Ctrl+Tab / Ctrl+Shift+Tab переключает табы
- [x] Ctrl+W закрывает текущий таб
- [x] Ctrl+T создаёт новый таб
- [x] Bubbletea-сообщения доставляются только сфокусированной панели
- [x] Q / Ctrl+C завершает программу (через tea.Quit)
- [x] Тесты проходят: `go test ./...`
- [x] `go vet ./...` без ошибок
- [x] Таб-бар настраиваемый (TabTop, TabBottom, TabLeft, TabRight)

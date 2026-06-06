# Warp — Handoff

## Что это

Warp — Go-библиотека (Bubbletea layout engine) для создания TUI с гибким управлением
пространством: вкладки, сплиты, плавающие панели, flexbox. Пользователь реализует интерфейс
`Panel`, а warp управляет их расположением.

## Состояние проекта

**v0.6** — TabGroup как Panel-компонент, tabs внутри splits/flex, 27 тестов.

## Новое в v0.6

- **TabGroup как Panel** — табы больше не глобальны. `TabGroup` реализует `Panel` и может
  быть вставлен внутри splits, flex layouts, или использован как root panel.
- **Warp — тонкая обёртка** — `Warp` теперь хранит `root Panel` и делегирует сообщения.
  `Warp.New()` создаёт root `TabGroup` для обратной совместимости.
- **Backward-compatible API** — `w.NewTab()`, `w.ActiveTab()`, `w.SetTabPosition()`
  делегируют root `TabGroup`.
- **Tab.handleMouse(cw, ch)** — теперь принимает размеры контента явно, не зависит от `Warp`.

## Демо (`cmd/demo/main.go`)

```bash
go run ./cmd/demo/
```

**Tab 1 — «main»**: FlexRow с 4 колонками:
- `Collapsible("Explorer")` — клик по ▶/▼ сворачивает
- `Scrollable` — длинный текст с WordWrap + Selectable, колесо мыши для скролла
- `DropdownMenu("Actions")` — клик открывает список
- `Collapsible("Terminal")` — ещё один сворачиваемый
- Float панель — drag за title bar, resize за края, × для закрытия

**Tab 2 — «nested»**: Вложенные табы (Warp внутри сплита)

**Tab 3 — «columns»**: FlexColumn с collapsible

**Tab 4 — «splits»**: Классические split-панели

**Горячие клавиши**:
- `Ctrl+T` — новый таб
- `Ctrl+W` — закрыть таб
- `Ctrl+Tab` / `Ctrl+Shift+Tab` — следующий/предыдущий таб
- `q` / `Ctrl+C` — выход

## Модуль и зависимости

- **Модуль:** `github.com/starframe-dev/wrap` (renamed from `github.com/Starframe/warp`)
- **Директория:** `/Users/a/Space/Projects/Starframe/warp`
- **Go:** 1.22+
- **Зависимости:** `bubbletea v1.1.0`, `lipgloss v0.13.0`
- **Тесты:** 27 тестов проходят, `go vet` чист.
- **Git:** `https://github.com/starframe-dev/wrap.git`, ветка `main`

## Архитектура

```
warp.go       — tea.Model, тонкая обёртка вокруг root Panel
tabgroup.go   — TabGroup: Panel с таб-баром, переключением табов, keyboard/mouse
tab.go        — Tab: дерево splits, float-панели, фокус, mouse handling, рендеринг контента
panel.go      — интерфейс Panel{View(w,h) string; Update(Msg) Cmd} + BasePanel
split.go      — Node, SplitConfig, Direction (Vertical/Horizontal), MinPanelSize=3
render.go     — renderNode (рекурсивный), findBorders (для drag hit-test), padContent
float.go      — FloatPane: рендеринг рамки, перемещение, ресайз, overlayFloat
styles.go     — lipgloss-стили для таб-бара, границ сплитов, float-рамок (Gruvbox Dark)
collapsible.go — Collapsible Panel с заголовком и toggle
scrollable.go  — Scrollable Panel с viewport и mouse wheel
dropdown.go    — DropdownMenu Panel с кнопкой и раскрывающимся списком
contextmenu.go — ContextMenu float Panel
selectable.go  — Selectable Panel с text selection (mouse drag, Shift+arrows, Ctrl+A)
wrap.go        — WordWrap, SpaceWrap утилиты для переноса текста
```

### Дерево панелей

Панели внутри Tab организованы в бинарное дерево:

```go
Node { Panel Panel | Split *SplitConfig | Flex *FlexConfig }
SplitConfig { Direction, Fraction, First *Node, Second *Node, Dragging bool }
FlexConfig  { Direction, Items []FlexItem }
```

Корень — всегда листовой узел с `emptyPanel` (если пользователь ничего не добавил).

### TabGroup как Panel

```go
// Табы как корень (как раньше)
w := warp.New()  // root = TabGroup с 1 табом

// Табы как компонент внутри flex
tg := warp.NewTabGroup(warp.TabTop)
tg.NewTab("code")
tg.NewTab("debug")
flex.Add(tg, 1)  // ← tabs inside flex!

// Warp вообще без табов
w := warp.New()
w.SetRoot(myCustomPanel)
```

### Рендеринг

- **Split** — рекурсивный: leaf → `panel.View(w, h)`, split → рендерим детей, склеиваем с границей `│`/`─`.
- **Flex** — распределение Grow-весов, drag resize через пересчёт Grow.
- **Float** — рендерятся поверх через `overlayFloat` — ANSI-aware наложение с `StripANSI()` и CSI-aware позиционированием.
- **Collapsible** — basis=1 когда Collapsed, иначе Grow.

### TabPosition

```go
TabTop    — таб-бар сверху (1 строка), контент снизу
TabBottom — таб-бар снизу (1 строка), контент сверху
TabLeft   — таб-бар слева (вертикальный), контент справа
TabRight  — таб-бар справа (вертикальный), контент слева
```

## API

```go
w := warp.New()
tab := w.NewTab("name")
w.SetTabPosition(warp.TabBottom)

// Layouts
tab.SplitVertical(parent, 0.5, newPanel)
tab.SplitHorizontal(parent, 0.5, newPanel)
tab.FlexRow(parent, []warp.FlexItemSpec{{Panel: p1, Grow: 1}, {Panel: p2, Grow: 2}})
tab.Float(panel, x, y, w, h)

// Collapsible panels
col := warp.NewCollapsible("Title", panel)
tab.FlexRow(parent, []warp.FlexItemSpec{{Panel: col, Grow: 1}})
tab.ToggleCollapsible(col)

// Scrollable content
scroll := warp.NewScrollable(panel) // mouse wheel / pgup/pgdown

// Dropdown menu
dd := warp.NewDropdownMenu("Menu", []warp.DropdownItem{
    {Label: "Item 1"}, {Label: "Item 2"},
})
dd.OnSelect = func(idx int) { /* ... */ }

// Context menu (right-click)
items := []warp.ContextMenuItem{
    {Label: "Copy", Shortcut: "Ctrl+C", Action: func() { /* ... */ }},
}
fp := tab.ShowContextMenu(items, x, y)

// Text selection
sel := warp.NewSelectable(panel)
selected := sel.SelectedText()
sel.ClearSelection()
sel.SelectAll(w, h)

// Word wrap
lines := warp.WordWrap(text, 40)      // break at word boundaries
lines := warp.SpaceWrap(text, 40)     // break at spaces only
wrapped := warp.WrapToString(text, 40, false)

// Nested warps
inner := warp.New()
inner.NewTab("inner")
tab.SplitVertical(tab.RootPanel(), 0.5, inner.AsPanel())

w.Run()
```

## Что не доделано / Ideas

- **Стилизация** — цвета захардкожены в styles.go, нет публичного API для кастомизации.
  lipgloss v0.13 требует `CLICOLOR_FORCE=1` для выдачи ANSI в не-TTY окружении.
- **Анимации** — нет (drag без анимации, переключение табов мгновенное)
- **Nested float** — float внутри float не поддерживается
- **Tab-close confirmation** — Ctrl+W закрывает без подтверждения
- **Context menu закрытие по клику вне** — нужно добавить глобальный mouse handler
- **OSC 52 clipboard** — копирование выделенного текста в системный clipboard

## Changelog

### v0.6
- **TabGroup как Panel** — табы внутри splits/flex, Warp — thin wrapper
- **Backward-compatible API** — `w.NewTab()`, `w.ActiveTab()` делегируют root TabGroup

### v0.5
- **Табы не прыгают** — фиксированная ширина вкладок
- **Scrollable** — `NewScrollable(panel)` с mouse wheel, pgup/pgdown
- **DropdownMenu** — `NewDropdownMenu(label, items)` с кнопкой ▼
- **ContextMenu** — `Tab.ShowContextMenu(items, x, y)`
- **WordWrap / SpaceWrap** — корректный перенос через `lipgloss.Width()`
- **Collapsible panels** — `NewCollapsible(title, panel)`
- **Flexbox** — `FlexRow()` / `FlexColumn()` с `FlexItemSpec{Panel, Grow}`
- **Gruvbox Dark тема**
- **Вложенные табы** — `Warp.AsPanel()`
- **WindowSizeMsg** — broadcast всем панелям
- **Float overlay fix** — ANSI-aware, не удлиняет строки
- **Float close button (×)** — детект клика, CloseRequested
- **24→27 тестов**

## Правила кода

- Не использовать `os.Exit()` — только `tea.Quit`
- Не добавлять свои `signal.Notify` — Bubbletea сам обрабатывает SIGINT
- Не shadow'ить receiver `w *Warp` в методах
- Минимальный размер панели: `MinPanelSize = 3`
- Fraction всегда через `clampFraction(0.1–0.9)`
- Файлы завершаются переносом строки
- Комментарии только на английском

## Как запустить тесты

```bash
cd /Users/a/Space/Projects/Starframe/warp
go vet ./...
go test ./... -v
```

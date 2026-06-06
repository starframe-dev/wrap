# Warp — Handoff

## Что это

Warp — Go-библиотека (Bubbletea layout engine) для создания TUI с гибким управлением
пространством: вкладки, сплиты, плавающие панели. Пользователь реализует интерфейс
`Panel`, а warp управляет их расположением.

## Состояние проекта

**v0.5** — табы не прыгают, scrollable, dropdown, context menu, word-wrap, 23 теста.

## Новое в v0.5

- **Табы не прыгают** — фиксированная ширина всех вкладок (активная и неактивные одинаковой ширины)
- **Scrollable** — `warp.NewScrollable(panel)` с mouse wheel, pgup/pgdown, up/down
- **DropdownMenu** — `warp.NewDropdownMenu(label, items)` с кнопкой ▼ и раскрывающимся списком
- **ContextMenu** — `tab.ShowContextMenu(items, x, y)` — float с меню по координатам
- **WordWrap / SpaceWrap** — корректный перенос с `lipgloss.Width()` (не байты)
- **Float overlay fix** — `overlayFloat` использует `lipgloss.Width()` для вычисления
  визуальной ширины float, а не `len()` (байты). UTF-8 символы (╭, ─, ╮ и т.д.)
  корректно обрабатываются — float больше не обрезает suffix и не удлиняет строки
- **Демо** — `cmd/demo/main.go` демонстрирует все возможности

## Демо (`cmd/demo/main.go`)

```bash
go run ./cmd/demo/
```

**Tab 1 — «main»**: FlexRow с 4 колонками:
- `Collapsible("Explorer")` — клик по ▶/▼ сворачивает
- `Scrollable` — длинный текст с WordWrap, колесо мыши для скролла
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

- **Модуль:** `github.com/Starframe/warp`
- **Директория:** `/Users/a/Space/Projects/Starframe/warp`
- **Go:** 1.22+
- **Зависимости:** `bubbletea v1.1.0`, `lipgloss v0.13.0`
- **Тесты:** 24 теста проходят, `go vet` чист.

## Архитектура

```
warp.go     — tea.Model, список табов, клавиатурная навигация, TabPosition
tab.go      — Tab: дерево splits, float-панели, фокус, mouse handling, рендеринг таб-бара
panel.go    — интерфейс Panel{View(w,h) string; Update(Msg) Cmd} + BasePanel
split.go    — Node, SplitConfig, Direction (Vertical/Horizontal), MinPanelSize=3
render.go   — renderNode (рекурсивный), findBorders (для drag hit-test), padContent
float.go    — FloatPane: рендеринг рамки, перемещение за заголовок, ресайз за края, overlayFloat
drag.go     — заглушка (drag-логика в tab.go: handleMouse → updateDrag)
styles.go   — lipgloss-стили для таб-бара, границ сплитов, float-рамок
```

### Дерево панелей

Панели внутри Tab организованы в бинарное дерево:

```go
Node { Panel Panel | Split *SplitConfig }
SplitConfig { Direction, Fraction, First *Node, Second *Node, Dragging bool }
```

Корень — всегда листовой узел с `emptyPanel` (если пользователь ничего не добавил).
При split листовой узел заменяется на SplitConfig с двумя детьми.

### Рендеринг

Рекурсивный: leaf → `panel.View(w, h)`, split → рендерим детей, склеиваем с границей `│`/`─`.
Float-панели рендерятся поверх через `overlayFloat` — ANSI-aware наложение,
корректно обрабатывающее CSI-последовательности (`\x1b[...m`).

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

// Word wrap
lines := warp.WordWrap(text, 40)      // break at word boundaries
lines := warp.SpaceWrap(text, 40)     // break at spaces only
wrapped := warp.WrapToString(text, 40, false)

w.Run()
```

## Что не доделано / Ideas

- **Стилизация** — цвета захардкожены в styles.go, нет публичного API для кастомизации.
  lipgloss v0.13 требует `CLICOLOR_FORCE=1` для выдачи ANSI в не-TTY окружении.
- **Анимации** — нет (drag без анимации, переключение табов мгновенное)
- **Nested float** — float внутри float не поддерживается
- **Tab-close confirmation** — Ctrl+W закрывает без подтверждения
- **Drag для flexbox** — resize через drag границ между flex items работает, но не оттестирован визуально
- **Context menu закрытие по клику вне** — нужно добавить глобальный mouse handler
- **Text selection** — выделение текста мышью внутри панелей

## Changelog (v0.5)

- **Табы не прыгают** — `renderHorizontalTabBar` вычисляет maxLabelW и pad-ит
  неактивные вкладки до той же ширины что активная (`▎ name ×` vs `  name  `)
- **Scrollable** — `NewScrollable(panel)` обёртка с viewport. Mouse wheel,
  pgup/pgdown, up/down. Авто-clamp offset
- **DropdownMenu** — `NewDropdownMenu(label, items)` с кнопкой ▼ и раскрывающимся
  списком. Hover, выбор enter/click, `OnSelect` callback
- **ContextMenu** — `Tab.ShowContextMenu(items, x, y)` создаёт float с меню.
  `ContextMenuItem{Label, Shortcut, Action}`. Hover + click/enter
- **WordWrap / SpaceWrap** — `WordWrap(text, width)` разрывает по границам слов,
  `SpaceWrap(text, width)` только по пробелам. `WrapToString()` convenience
- **Collapsible panels** — `NewCollapsible(title, panel)` + `Tab.ToggleCollapsible()`
- **Flexbox** — `FlexRow()` / `FlexColumn()` с `FlexItemSpec{Panel, Grow}`
- **Gruvbox Dark тема**
- **Вложенные табы** — `Warp.AsPanel()`
- **WindowSizeMsg** — broadcast всем панелям
- **24 теста:** включая float overlay, tab alignment, scrollable, word-wrap, flex, collapsible

## Правила кода

- Не использовать `os.Exit()` — только `tea.Quit`
- Не добавлять свои signal.Notify — Bubbletea сам обрабатывает SIGINT
- Не shadow'ить receiver `w *Warp` в методах (уже исправлено в renderVerticalTabBar)
- Минимальный размер панели: MinPanelSize = 3
- Fraction всегда через clampFraction (0.1–0.9)

## Связанные проекты

- `HumanHorizon/automata` — AI-агент, для которого warp создаётся
- `Lattice/tui-launcher/viewer` — терминальный мультиплексор, референс архитектуры
  (Bubbletea + вкладки + unix socket, но с VT10x/PTY)

## Как запустить тесты

```bash
cd /Users/a/Space/Projects/Starframe/warp
go vet ./...
go test ./... -v
```

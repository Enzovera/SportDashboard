# Design — kierunek v2

Makiety w `screens/` to statyczny eksport z makiet interaktywnych (https://claude.ai/artifact/KtXo6nwS3hwJma1m7nRzpi). Każdy ekran ma wersję jasną (`X.html`) i ciemną (`X-dark.html`); linki w nawigacji prowadzą między plikami. Podgląd wszystkich: `index.html`. Zrzuty PNG (2×) w `png/`.

## Ekrany

| Plik | Ekran | Rozmiar |
|---|---|---|
| `FoldDzis` | Dziś — podsumowanie trenera, sesja dnia (hero), gotowość, żywienie, tydzień | 707 × 823 |
| `FoldTrening` | Trening — obciążenie (CTL/ATL/TSB), kilometraż, ostatnie biegi, serie PPL | 707 × 1040 |
| `FoldZywienie` | Żywienie — bilans, kalorie dzień po dniu, białko g/kg, cele per typ dnia, posiłki | 707 × 850 |
| `FoldSuplementy` | Żywienie › Suplementy — checklista dnia, regularność, zapasy, propozycja trenera | 707 × 823 |
| `FoldCialo` | Ciało — skład ciała, waga tydzień po tygodniu, ważenia | 707 × 823 |
| `FoldPlan` | Plan — cel (hero), oś makrocyklu, następny tydzień z propozycjami | 707 × 823 |
| `FoldPropozycje` | Propozycje — karty zmian z walidatorami, Akceptuj / Odrzuć / Cofnij, historia | 707 × 960 |
| `FoldCzat` | Plan + czat z trenerem — zmiana z czatu widoczna w planie | 707 × 823 |
| `CoverDzis` | Dziś na ekranie zewnętrznym, z checklistą suplementów | 369 × 1400 |
| `CoverPropozycje` | Propozycje na ekranie zewnętrznym | 369 × 1650 |
| `CoverCzat` | Czat z trenerem na ekranie zewnętrznym | 369 × 905 |

Wysokość powyżej 823 (lub 905 na zewnętrznym) oznacza przewijanie.

## Zasady

- **Kształt:** zero zaokrągleń (`border-radius: 0` wszędzie, także przyciski, plakietki, pola wyboru). Boksy: tło `surface`, ramka 1px `border`, padding 12–14 px. Bez cieni.
- **Hierarchia:** jeden dominujący blok na ekran (`hero`: ciemne tło), reszta to równe boksy. Tytuły sekcji 14 px / 700, meta 12.5 px w `ink3` po prawej.
- **Typografia:** Archivo (oś szerokości). Duże liczby wąsko (`font-stretch` 70–78%, waga 750–800), wszędzie `font-variant-numeric: tabular-nums`. Jednostki mniejsze i w `ink2`.
- **Kolor:** interfejs czarno-biały. `signal` (czerwień) wyłącznie dla „dziś / teraz / wymaga decyzji”: dzisiejszy dystans, dzisiejsza kolumna, licznik propozycji, wiersze do decyzji. Ostrzeżenia i błędy na `warn` (bursztyn), stan w normie na `good`. Akcja główna: tło `ink`, tekst `surface`.
- **Dane:** tabele zamiast pasków postępu, gdy liczba wystarcza. Wykresy: seria główna `ink`, wtórna `chartMid`, tło/odciążenie `chartSoft`, punkty kwadratowe, „dziś” na czerwono.
- **Tekst:** bez mikrotekstów tłumaczących działanie aplikacji („trener zmienia tylko…”, „Agent SDK…”). Zamiast objaśnień — przyciski akcji. Proza trenera (LLM) jest w porządku.
- **Nawigacja:** ikona nad etykietą; aktywna zakładka pogrubiona z kreską 3 px u góry, licznik propozycji na ikonie. Te same sześć zakładek na obu ekranach: na Fold6 dwie grupy 3 | 3, na zewnętrznym jeden rząd.
- **Ikony:** liniowe na siatce 24 px, kreska 1,75 px, kwadratowe zakończenia i ostre narożniki (`web/src/components/ui/Icon.vue`). Akcje nagłówka na Fold6 i ekranie zewnętrznym to same ikony 44 × 44 px (Odśwież, Trener, motyw) z nazwą w `aria-label`; na desktopie ikona z etykietą. Strzałki w wierszach-linkach też są ikonami.
- **Nagłówek na ekranie zewnętrznym:** tytuł nad linią meta (tam też status syncu), ikony po prawej. Przełączniki widoku (zakres w Treningu, Jedzenie | Suplementy) schodzą do paska zakładek pod nagłówkiem: równe kolumny, aktywna z kreską 3 px u dołu.
- **Fold6:** ekran wewnętrzny 707 × 823, dwie kolumny (margines 16, odstęp 24 — zgięcie wypada w odstępie), pasek górny 56 px, nawigacja 58 px (plus `safe-area-inset`). Ekran zewnętrzny 369 px, jedna kolumna, margines 12.
- **Dostępność:** cele dotykowe ≥ 44 px, kontrast tekstu ≥ 4.5:1 w obu motywach, statusy zawsze także słowem (nie tylko kolorem).

Tokeny: `tokens.css` (zmienne CSS, motyw przez `prefers-color-scheme` lub `data-theme`) i `tokens.json` (te same wartości z opisem użycia).

## Desktop (web)

Makiety `screens/Desktop*.html` (+ `-dark`) dla przeglądarki od 1024 px, projekt na 1440 × 900. Zasady, które różnią się od Fold6, są w `desktop.md`; wspólne style w `screens/desktop.css`. Otwieraj przez serwer makiet (`http://localhost:8082`), bo strony ładują wspólny CSS. PNG dla desktopu jeszcze nie ma.

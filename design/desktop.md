# Design — wersja webowa (desktop), v2

Rozszerzenie zasad z `design/README.md` na przeglądarkę na komputerze. Wszystkie zasady v2 obowiązują bez zmian (zero zaokrągleń, boksy z ramką 1px, Archivo z wąskimi liczbami, czerwień tylko dla „dziś / teraz / wymaga decyzji”, bursztyn dla ostrzeżeń, słowo zawsze obok koloru, tabele zamiast pasków, brak mikrotekstów). Ten plik opisuje tylko to, co desktop robi inaczej.

## Kiedy desktop

- Układ desktopowy od **1024 px** szerokości okna. Poniżej działa układ Fold6 (707) i zewnętrzny (369).
- Projektujemy na **1440 × 900**; ekran musi się też mieścić bez poziomego przewijania przy **1280 px**. Wysokość może przewijać — nagłówek zostaje na górze.

## Powłoka

| Element | Desktop | Fold6 (dla porównania) |
|---|---|---|
| Nawigacja | w nagłówku, tekstowa, 6 zakładek w jednym rzędzie, aktywna pogrubiona z kreską 3 px **u dołu** nagłówka | dolny pasek 3 \| 3, kreska u góry |
| Nagłówek | 64 px: znak „Forma”, nawigacja, po prawej status syncu i przycisk Trener | 56 px: tytuł ekranu i akcje |
| Tytuł ekranu | w treści, nad siatką: h1 32 px + meta 14 px, akcje ekranu (np. zakres 4 tyg. / 12 tyg. / Blok) po prawej | w nagłówku |
| Siatka | 12 kolumn, odstęp 24, treść max 1392 px, margines 32 | 2 kolumny, odstęp 24, margines 16 |
| Odstęp między boksami w kolumnie | 16 | 12 |
| Czat z trenerem | panel po prawej, 420 px, cała wysokość pod nagłówkiem; treść się zwęża, nie jest przykrywana | osobna kolumna zamiast prawej (FoldCzat) |

## Skala

Desktop czyta się z większej odległości niż telefon w dłoni, więc wszystko jest trochę większe, ale proporcje zostają:

- tekst bazowy 14–15 px, proza trenera 16 px / 1.55
- tytuły boksów 15 px / 700, meta 13 px `ink3`
- duże liczby w siatce metryk 40 px (Fold: 34), hero 96 px (Fold: 80)
- padding boksu 16 × 18 (Fold: 12 × 14)
- cele dotykowe i kliknięcia ≥ 44 px (to samo)

## Kompozycja

- Jeden dominujący blok (`hero`) na ekran, jak na Fold6.
- Szerokość daje miejsce, nie powód do dokładania treści: te same sekcje co na Fold6, ułożone tak, żeby **najważniejsze było w pierwszym ekranie (900 px) bez przewijania**.
- Wykresy mogą być szersze (6–8 kolumn) i wyższe (240–280 px), ale z tymi samymi kolorami serii, kwadratowymi punktami i „dziś” na czerwono.
- Siatka reaguje na szerokość kolumny treści (container query na `.d-main`): przy ≤ 1100 px, czyli z otwartym czatem albo w wąskim oknie, każdy boks zajmuje 6 kolumn.
- Hero tylko tam, gdzie jedna rzecz dominuje: Dziś (sesja), Plan (cel), Żywienie (dziś vs cel), Suplementy (checklista), Czat (jak Dziś). Trening, Ciało i Propozycje nie mają hero: tam kotwicą jest szeroki wykres albo równorzędne karty do decyzji. Na tle hero stany w normie i ostrzeżenia mają własne tokeny `--hero-good` i `--hero-warn`.
- Wiersze „do decyzji” (`tr.row-decide`: tło `signal-tint` + pasek 3 px `signal`) znaczą to samo w Planie, w diffie tygodnia w Propozycjach i w karcie propozycji w czacie.
- Wykresy z tej samej osi czasu, ustawione jeden pod drugim, dzielą pozycje X (np. kcal i białko dzień po dniu). Wykres w s-8 rysuj w viewBox 840 × 240–280 (skala 1:1 przy 1440); etykiety końcowe serii w prawym marginesie.
- Log czatu jest przewinięty do najnowszej wiadomości; karta propozycji z Odrzuć / Akceptuj jest ostatnim elementem.
- Tabele mogą pokazać więcej kolumn niż na Fold6, jeśli dane są w kontrakcie API (`web/src/api/types.ts`) albo wynikają z makiet Fold6. Nie wymyślamy nowych metryk.

## Pliki

- `design/screens/desktop.css` — powłoka i prymitywy (te same nazwy klas co `web/src/styles/base.css`, żeby makieta przekładała się 1:1 na kod).
- `design/screens/DesktopTemplate.html` — szablon: nagłówek, tytuł, siatka.
- `design/screens/Desktop<Ekran>.html` + `Desktop<Ekran>-dark.html` (ciemny = ten sam plik z `<html data-theme="dark">`).

| Plik | Ekran |
|---|---|
| `DesktopDzis` | Dziś |
| `DesktopCzat` | Dziś z otwartym panelem czatu |
| `DesktopTrening` | Trening |
| `DesktopCialo` | Ciało |
| `DesktopZywienie` | Żywienie |
| `DesktopSuplementy` | Żywienie › Suplementy |
| `DesktopPlan` | Plan |
| `DesktopPropozycje` | Propozycje |

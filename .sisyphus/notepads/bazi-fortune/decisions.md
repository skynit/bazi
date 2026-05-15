# Decisions - Bazi Fortune Frontend

## ChartView dual-mode (new vs view)
ChartView handles both `/chart/new` (shows BirthInputForm) and `/chart/:id` (shows BaziChart). This avoids a separate route for new chart creation.

## BaziChart element coloring
- Gan stems use Tailwind text color classes (text-green-700, text-red-600, etc.)
- Zhi branches also color-coded by five-element
- Clash relationships shown as text badges rather than graphical lines for clarity
- "金" element uses amber-600 (golden) since white doesn't work on light backgrounds

## Auth store persistence
Token persisted to localStorage only. No refresh token mechanism (simple approach). User object fetched on HomeView mount if logged in.

## API client 401 handling
On 401 response, clears token from localStorage AND redirects to /login. This covers both expired tokens and unauthorized access.

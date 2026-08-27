# کلنگ-لینتر — لینتر زبان برنامه‌نویسی فارسی کلنگ

> لینتر مستقل زبان برنامه‌نویسی فارسی کلنگ؛ بررسی کد کلنگ و خروجی به‌صورت JSON

[![CI](https://img.shields.io/badge/CI-passing-brightgreen)](https://github.com/faralidev/kolang-linter/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/faralidev/kolang-linter.svg)](https://pkg.go.dev/github.com/faralidev/kolang-linter)
![Go](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

**[نسخهٔ انگلیسی در پایین صفحه](#english-section)**

---

## ۱. معرفی

کلنگ-لینتر یک ابزار مستقل برای بررسی و لینت کدهای زبان برنامه‌نویسی فارسی [کلنگ](https://github.com/faralidev/kolang) است. این ابزار کد کلنگ را از یک فایل یا از ورودی استاندارد (stdin) می‌خواند و نتیجهٔ بررسی را به‌صورت JSON در خروجی استاندارد (stdout) می‌نویسد. خروجی این لینتر برای مصرف در ویرایشگرهای کد (از جمله ویرایشگر کلنگ مبتنی بر CodeMirror 6) طراحی شده است.

کلنگ-لینتر ۱۳ قاعدهٔ بررسی را در سه دستهٔ **نحوی**، **سبکی** و **معنایی** اجرا می‌کند؛ پیام‌های همهٔ قواعد به فارسی و ساده هستند.

---

## ۲. نصب

### روش ۱: نصب با Homebrew

```bash
brew install faralidev/tap/kolang-linter
```

### روش ۲: نصب با Go

```bash
go install github.com/faralidev/kolang-linter@latest
```

### روش ۳: ساخت از سورس

```bash
git clone https://github.com/faralidev/kolang-linter.git
cd kolang-linter
go build -o kolang-linter .
```

**پیش‌نیاز:** Go نسخهٔ ۱.۲۷ یا بالاتر (برای روش‌های ۲ و ۳).

---

## ۳. استفاده

### بررسی یک فایل

```bash
kolang-linter file.kolang
```

### ورودی از استاندارد

```bash
echo '«سلام» بنویس' | kolang-linter
```

### گزینه‌های خط فرمان

| گزینه | توضیح |
|-------|-------|
| `-format json` | قالب خروجی؛ تنها قالب پشتیبانی‌شده در نسخهٔ ۱ (پیش‌فرض) |
| `-strict` | برای سازگاری رابط خط فرمان پذیرفته می‌شود؛ در نسخهٔ ۱ رفتاری ندارد |

**کد خروجی:** در صورت اجرای موفق (حتی اگر تشخیصی گزارش شود) کد `0` و در صورت خطای داخلی (فایل ناخوانا یا گزینهٔ نامعتبر) کد `1` برگردانده می‌شود.

### مثال

فرض کنید فایل `example.kolang` این‌گونه باشد:

```kolang
سن = ۱۸
اگر سن:
    «بزرگسال» بنویس
```

اجرای دستور:

```bash
kolang-linter example.kolang
```

خروجی:

```json
{
  "diagnostics": [
    {
      "line": 2,
      "col": 1,
      "endLine": 2,
      "endCol": 2,
      "severity": "error",
      "message": "شرط باید شامل مقایسه و باشد/نباشد باشد — مقایسه ضمنی مجاز نیست",
      "rule": "no-implicit-truthiness"
    }
  ]
}
```

---

## ۴. ساختار خروجی

خروجی لینتر یک سند JSON با کلید `diagnostics` است که آرایه‌ای از تشخیص‌ها را نگه می‌دارد. هر تشخیص شامل این فیلدهاست:

| فیلد | توضیح |
|------|-------|
| `line` | شمارهٔ خط شروع (یک‌مبنایی) |
| `col` | شمارهٔ ستون شروع (یک‌مبنایی) |
| `endLine` | شمارهٔ خط پایان (یک‌مبنایی) |
| `endCol` | شمارهٔ ستون پایان — **انحصاری** (یکی بعد از آخرین نویسه) |
| `severity` | شدت تشخیص: `error`، `warning` یا `info` |
| `message` | پیام فارسی تشخیص |
| `rule` | شناسهٔ قاعدهٔ مربوط |

برای کد سالم، خروجی به این شکل است:

```json
{"diagnostics":[]}
```

---

## ۵. قواعد

کلنگ-لینتر ۱۳ قاعده را در سه دسته اجرا می‌کند. شدت‌ها مطابق مقادیر JSON (خطا، هشدار، اطلاع) گزارش می‌شوند.

### قواعد نحوی

| شناسهٔ قاعده | شدت | رفتار |
|--------------|------|-------|
| `syntax-error` | `error` | خطای نحو در زمان تجزیهٔ برنامه؛ شمارهٔ خط از پیام خطای تجزیه‌کننده استخراج می‌شود |
| `unclosed-string` | `error` | متن با « شروع‌شده اما بدون » پایانی؛ توسط تحلیل واژگانی علامت‌گذاری می‌شود |
| `unclosed-comment` | `warning` | کامنت بلوکی `//` که با `//` دوم بسته نشده است |

### قواعد سبکی

| شناسهٔ قاعده | شدت | رفتار |
|--------------|------|-------|
| `no-implicit-truthiness` | `error` | شرط بدون عملگر مقایسه و بدون «باشد/نباشد» (مثلاً `اگر x:`)— بر اساس مشخصات کلنگ، مقایسهٔ ضمنی مجاز نیست |
| `negation-no-bang-eq` | `error` | استفاده از عملگر `!=` که از زبان حذف شده؛ باید از `== ... نباشد` استفاده کرد |
| `dot-access` | `error` | دسترسی به عضو با نقطه (`a.b`)؛ باید از اضافه (ِ) استفاده کرد — نقطهٔ جداکنندهٔ اعشار نادیده گرفته می‌شود |
| `line-too-long` | `warning` | خط‌های طولانی‌تر از ۱۰۰ نویسه |
| `mixed-indentation` | `warning` | ترکیب فاصله (space) و تب (tab) در تورفتگی یک خط |
| `trailing-whitespace` | `info` | وجود فاصله یا تب در انتهای خط |

### قواعد معنایی

| شناسهٔ قاعده | شدت | رفتار |
|--------------|------|-------|
| `undefined-variable` | `warning` | استفاده از شناسه‌ای که در فایل تعریف نشده است؛ توابع از پیش‌تعریف‌شدهٔ کلنگ، نام استثناها و جایگاه‌های نام عضو نادیده گرفته می‌شوند |
| `unused-variable` | `warning` | متغیری که مقداردهی شده اما هیچ‌جا خوانده نمی‌شود (فقط متغیرها؛ نه توابع، کلاس‌ها و ماژول‌ها) |
| `naming-convention` | `info` | ترکیب حروف لاتین و فارسی در یک نام، مثلاً `myVarنام` |
| `duplicate-import` | `warning` | وارد کردن دوبارهٔ یک ماژول |

---

## ۶. توسعه

اجرای تست‌ها:

```bash
go test ./...
```

ساخت فایل اجرایی:

```bash
go build -o kolang-linter .
```

---

## ۷. مجوز

این پروژه تحت مجوز MIT منتشر شده است.

---

# English Section

# kolang-linter — Linter for the Persian Kolang Language

> A standalone linter for the Kolang (کلنگ) Persian programming language.

kolang-linter reads Kolang source from a file argument or stdin and emits JSON diagnostics to stdout. Its output is designed for consumption by a Persian-language IDE (CodeMirror 6).

## Installation

```bash
brew install faralidev/tap/kolang-linter   # Homebrew
go install github.com/faralidev/kolang-linter@latest
```

Or build from source: `go build -o kolang-linter .`

## Usage

```bash
kolang-linter file.kolang   # lint a file
echo '«سلام» بنویس' | kolang-linter   # lint stdin
kolang-linter -format json file.kolang   # explicit JSON format (the default)
```

- Exit code `0` on a successful run even when diagnostics are reported; `1` on an internal error.
- Output is `{"diagnostics":[...]}`, with each diagnostic carrying `line`, `col`, `endLine`, `endCol`, `severity`, `message`, and `rule`. Positions are 1-based; `endCol` is exclusive. Valid code yields `{"diagnostics":[]}`.

## Rules

13 rules in three categories:

- **Syntax:** `syntax-error` (error), `unclosed-string` (error), `unclosed-comment` (warning)
- **Style:** `no-implicit-truthiness` (error), `negation-no-bang-eq` (error), `dot-access` (error), `line-too-long` (warning), `mixed-indentation` (warning), `trailing-whitespace` (info)
- **Semantic:** `undefined-variable` (warning), `unused-variable` (warning), `naming-convention` (info), `duplicate-import` (warning)

All messages are in Persian.

## Development

```bash
go test ./...
go build -o kolang-linter .
```

## License

MIT.
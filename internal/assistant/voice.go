package assistant

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"
)

// ── STT / TTS / Wake word — semua lewat interface, implementasi konkret ──
// (worker Python) ada di file terpisah supaya package ini tidak perlu tahu
// SenseVoice/Kokoro sama sekali — persis prinsip "Go tidak tahu model apa,
// Go cuma tahu Generate Speech" yang kamu maksud.

type SpeechToText interface {
	// Transcribe: kirim audio (Opus frame per elemen), tunggu transkrip final.
	Transcribe(ctx context.Context, opusFrames [][]byte, conversationID uuid.UUID) (string, error)
}

type TextToSpeech interface {
	// SynthesizeStream: teks masuk, Opus packet (raw bytes, BUKAN base64 —
	// base64-nya dilakukan pas nulis ke WS Flutter, bukan di sini) keluar
	// lewat channel. Channel ditutup otomatis saat sintesis kalimat itu selesai.
	SynthesizeStream(ctx context.Context, text string, personality Personality, conversationID uuid.UUID) (<-chan []byte, error)
}

type WakeWordProvider interface {
	GetWakeWords(ctx context.Context, deviceID string) (map[string]Personality, error)
}

// ── Wake word check ──────────────────────────────────────────────────────

type WakeWordChecker struct {
	STT      SpeechToText
	Provider WakeWordProvider
}

func (c *WakeWordChecker) Check(ctx context.Context, deviceID string, opusFrames [][]byte, conversationID uuid.UUID) (matched bool, personality Personality, err error) {
	if c == nil || c.STT == nil || c.Provider == nil {
		return false, "", fmt.Errorf("wake word checker belum dikonfigurasi")
	}
	if len(opusFrames) == 0 {
		return false, "", nil
	}

	text, err := c.STT.Transcribe(ctx, opusFrames, conversationID)
	if err != nil {
		return false, "", fmt.Errorf("wake check transcribe: %w", err)
	}
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return false, "", nil
	}

	wakeWords, err := c.Provider.GetWakeWords(ctx, deviceID)
	if err != nil {
		return false, "", fmt.Errorf("wake check load wake words: %w", err)
	}

	for word, p := range wakeWords {
		if strings.HasPrefix(text, strings.ToLower(strings.TrimSpace(word))) {
			return true, p, nil
		}
	}
	return false, "", nil
}

// ── Sentence Builder ──────────────────────────────────────────────────────
//
// Terima token/delta text satu-satu (atau full text sekaligus lewat Feed
// besar + Flush), keluarkan kalimat LENGKAP begitu ketemu batas kalimat
// (. ! ? atau newline), supaya bisa langsung dilempar ke TTS tanpa nunggu
// LLM selesai total. Ini komponen murni (no I/O), jadi bisa dipakai baik
// untuk mode true token-streaming maupun fallback split-after-the-fact.
type SentenceBuilder struct {
	buf strings.Builder
}

// sentenceEnders: tanda baca yang menandakan akhir kalimat.
var sentenceEnders = []rune{'.', '!', '?', '\n'}

// minSentenceRunes: jangan split kalimat yang KETERLALUAN pendek (mis. "Rp.
// 500,-" ke-split di titik singkatan) — heuristik sederhana, bukan NLP proper.
const minSentenceRunes = 6

// Feed menambahkan potongan teks baru, mengembalikan kalimat-kalimat yang
// SUDAH lengkap (bisa 0, 1, atau lebih kalau delta-nya besar).
func (b *SentenceBuilder) Feed(delta string) []string {
	b.buf.WriteString(delta)
	return b.extractComplete(false)
}

// Flush dipanggil saat LLM/teks sumber sudah benar-benar habis — sisa
// buffer (walau belum diakhiri tanda baca) dikeluarkan sebagai kalimat
// terakhir.
func (b *SentenceBuilder) Flush() []string {
	return b.extractComplete(true)
}

func (b *SentenceBuilder) extractComplete(final bool) []string {
	var out []string
	content := b.buf.String()

	for {
		idx := indexAny(content, sentenceEnders)
		if idx == -1 {
			break
		}
		candidate := strings.TrimSpace(content[:idx+1])
		rest := content[idx+1:]

		if len([]rune(candidate)) < minSentenceRunes && !final {
			// Kependekan (kemungkinan singkatan) & bukan akhir stream —
			// tunggu token berikutnya sebelum memutuskan ini akhir kalimat.
			break
		}
		if candidate != "" {
			out = append(out, candidate)
		}
		content = rest
	}

	b.buf.Reset()
	b.buf.WriteString(content)

	if final {
		last := strings.TrimSpace(b.buf.String())
		if last != "" {
			out = append(out, last)
		}
		b.buf.Reset()
	}
	return out
}

func indexAny(s string, chars []rune) int {
	for i, r := range s {
		for _, c := range chars {
			if r == c {
				return i
			}
		}
	}
	return -1
}

// StreamingChatter: interface TAMBAHAN opsional untuk llm.Provider. Kalau
// provider LLM yang kamu pakai (internal/llm) sudah/nanti mengimplementasi
// ini, Stream() otomatis pakai TRUE token-by-token streaming ke Sentence
// Builder (efek JARVIS penuh). Kalau belum, Stream() fallback ke mode
// non-streaming (tunggu balasan penuh, baru displit per kalimat) — tetap
// dapat audio incremental, cuma TTS-nya baru mulai setelah LLM 100% selesai.
//
// TODO: aku belum lihat isi internal/llm, jadi cek dulu apakah provider yang
// kamu pakai sekarang bisa/gampang diberi kemampuan ini.
type StreamingChatter interface {
	ChatStream(ctx context.Context, req llm.ChatRequest) (<-chan string, error)
}

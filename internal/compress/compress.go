package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}
func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.New(os.Stdout)
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Warn().Msg(err.Error())
				return
			}
			defer gz.Close()
			r.Body = io.ReadCloser(gz)
		}
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}
		//todo: добавить ограничения по типу файлов и размеру (?проверить по тестам)
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			log.Warn().Msg(err.Error())
			return
		}
		//todo: как отказаться от постоянного вызова gzip.NewWriterLevel и использовать метод gzip.Reset, чтобы избежать выделения памяти при каждом запросе.
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)

	})
}

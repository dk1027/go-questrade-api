package controlflow

import (
	"bufio"
	"bytes"
	"fmt"
	"text/tabwriter"
)

func ToText(headers []string, rows []Table) string {
	const padding = 3
	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)
	w := tabwriter.NewWriter(writer, 0, 0, padding, ' ', tabwriter.Debug)
	s := ""
	for _, v := range headers {
		s = fmt.Sprintf("%v\t%v", s, v)
	}
	_, _ = fmt.Fprintln(w, s)
	for _, r := range rows {
		s = ""
		for _, k := range headers {
			s = fmt.Sprintf("%v\t%.2f", s, r[k])
		}
		_, _ = fmt.Fprintln(w, s)
	}

	_ = w.Flush()
	_ = writer.Flush()
	return buff.String()
}

package checksdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWordInExpr(t *testing.T) {
	type args struct {
		words []string
		expr  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no words",
			args: args{
				words: []string{},
				expr:  "word1",
			},
			want: false,
		},
		{
			name: "simple word 1: match",
			args: args{
				words: []string{"word1"},
				expr:  "word1",
			},
			want: true,
		},
		{
			name: "simple word 2: no match",
			args: args{
				words: []string{"word1"},
				expr:  "word2",
			},
			want: false,
		},
		{
			name: "word1 matches in OR expr",
			args: args{
				words: []string{"word1"},
				expr:  "word2 || word1",
			},
			want: true,
		},
		{
			name: "word1 matches in OR expr",
			args: args{
				words: []string{"word1"},
				expr:  "word2 || word1 || word3",
			},
			want: true,
		},
		{
			name: "word1 matches in OR and AND expr",
			args: args{
				words: []string{"word1"},
				expr:  "(word2 && word3) || word1 ",
			},
			want: true,
		},
		{
			name: "word2 && word3 matches in OR and AND expr",
			args: args{
				words: []string{"word2", "word3"},
				expr:  "word1 || (word2 && word3)",
			},
			want: true,
		},
		{
			name: "word2 won't match in OR and AND expr",
			args: args{
				words: []string{"word2"},
				expr:  "word1 || (word2 && word3)",
			},
			want: false,
		},
		{
			name: "word2 won't match in OR and AND expr",
			args: args{
				words: []string{"word1-word2"},
				expr:  "word1-word2 || (word2 && word3)",
			},
			want: true,
		},
		{
			name: "word2 won't match in OR and AND expr",
			args: args{
				words: []string{"word1-word2"},
				expr:  "word1 || (word2 && word3)|| word1-word2",
			},
			want: true,
		},
		{
			name: "word2 won't match in OR and AND expr",
			args: args{
				words: []string{"word1-word2", "word1"},
				expr:  "word1 && (word1-word2 || word3)",
			},
			want: true,
		},
		{
			name: "word2 won't match in OR and AND expr",
			args: args{
				words: []string{"word1-word2", "word1"},
				expr:  "word1, word3",
			},
			want: true,
		},
		{
			name: "word1 and not workd2",
			args: args{
				words: []string{"word2", "word1"},
				expr:  "word1 && !word3",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get labels expr evaluator
			labelsExprEvaluator, err := NewLabelsExprEvaluator(tt.args.expr)
			assert.NotNil(t, labelsExprEvaluator)
			assert.Nil(t, err)

			got := labelsExprEvaluator.Eval(tt.args.words)
			assert.Equal(t, got, tt.want)
		})
	}
}

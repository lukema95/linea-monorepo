package symbolic_test

import (
	"fmt"
	"path"
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/backend/files"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/mempool"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/serialization"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

func BenchmarkEvaluationSingleThreaded(b *testing.B) {

	makeRegular := func() smartvectors.SmartVector {
		return smartvectors.Rand(symbolic.MaxChunkSize)
	}

	makeConst := func() smartvectors.SmartVector {
		var x field.Element
		x.SetRandom()
		return smartvectors.NewConstant(x, symbolic.MaxChunkSize)
	}

	makeFullZero := func() smartvectors.SmartVector {
		return smartvectors.NewConstant(field.Zero(), symbolic.MaxChunkSize)
	}

	makeFullOnes := func() smartvectors.SmartVector {
		return smartvectors.NewConstant(field.One(), symbolic.MaxChunkSize)
	}

	for ratio := 1; ratio <= 32; ratio *= 2 {

		b.Run(fmt.Sprintf("ratio-%v", ratio), func(b *testing.B) {

			var (
				testDir           = "testdata/evaluation-benchmark"
				constanthoodFName = fmt.Sprintf("global-variable-constanthood-%v.csv", ratio)
				exprFName         = fmt.Sprintf("global-cs-ratio-%v.cbor.gz", ratio)
				constanthoodFPath = path.Join(testDir, constanthoodFName)
				exprFPath         = path.Join(testDir, exprFName)
				constantHoodFile  = files.MustRead(constanthoodFPath)
				exprFile          = files.MustReadCompressed(exprFPath)
				constantHood      = symbolic.ReadConstanthoodFromCsv(constantHoodFile)
				expr              = serialization.UnmarshalExprCBOR(exprFile)
				inputs            = make([]smartvectors.SmartVector, len(constantHood))
				board             = expr.Board()
				pool              = mempool.Create(symbolic.MaxChunkSize)
			)

			for i := range inputs {
				switch {
				case !constantHood[i][0]:
					inputs[i] = makeRegular()
				case constantHood[i][1]:
					inputs[i] = makeFullZero()
				case constantHood[i][2]:
					inputs[i] = makeFullOnes()
				default:
					inputs[i] = makeConst()
				}
			}

			b.ResetTimer()

			for c := 0; c < b.N; c++ {
				_ = board.Evaluate(inputs, pool)
			}
		})

	}
}

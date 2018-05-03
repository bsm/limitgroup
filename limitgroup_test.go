package limitgroup_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bsm/limitgroup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Group", func() {
	var subject *limitgroup.Group
	var ctx context.Context

	BeforeEach(func() {
		subject, ctx = limitgroup.New(context.Background(), 2)
	})

	It("should limit", func() {
		start := time.Now()
		subject.Go(func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		})
		subject.Go(func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		})
		subject.Go(func() error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		Expect(ctx.Done()).NotTo(BeClosed())
		Expect(subject.Wait()).To(Succeed())
		Expect(time.Since(start)).To(BeNumerically("~", 250*time.Millisecond, 30*time.Millisecond))
		Expect(ctx.Done()).To(BeClosed())
	})

	It("should fail on first error", func() {
		err := errors.New("stopped")
		inc := int64(0)
		increment := func(sleep time.Duration, by int64) func() error {
			return func() error {
				time.Sleep(sleep)
				atomic.AddInt64(&inc, by)
				return nil
			}
		}

		start := time.Now()
		for i := 0; i < 10; i++ {
			delay := time.Duration(i) * 10 * time.Millisecond
			subject.Go(increment(delay, 1))
		}
		subject.Go(func() error { return err })
		subject.Go(increment(0, 3))
		subject.Go(increment(0, 5))
		subject.Go(increment(0, 7))
		Expect(subject.Wait()).To(MatchError(err))
		Expect(time.Since(start)).To(BeNumerically("~", 250*time.Millisecond, 30*time.Millisecond))
		Expect(ctx.Done()).To(BeClosed())
		Expect(atomic.LoadInt64(&inc)).To(Equal(int64(13)))
	})

})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "limitgroup")
}

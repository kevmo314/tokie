# tokie

Tokie is a pure-Go byte pair encoding tokenizer.

An equivalent library is [OpenAI's tiktoken](https://github.com/openai/tiktoken).

I wrote this library with the intent of trying to beat tiktoken in performance. It turns
out the majority of performance is dominated by regexp operations. The good news is the
library is able to generally match tiktoken in performance and I'd like to think my
encoding algorithm is slightly more clever than the one in OpenAI's library, but this
ended up being less fruitful than I had hoped.

Probably the best way to fix this would be to remove the regexp layer and do the pattern splitting
in the same pass as the encoding however this is both quite involved and would be very hard
to adopt to different models so it's not really practical.

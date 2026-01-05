#include <stdint.h>
#include "../../third_party/libaesrand/aesrand_buf.c"
#include "../../third_party/libaesrand/aesrand_gmp.c"
#include "../../third_party/libaesrand/aesrand_init.c"
#include "../../third_party/clt13/src/utils.c"
#include "../../third_party/clt13/src/crt_tree.c"
#include "../../third_party/clt13/src/estimates.c"
#include "../../third_party/clt13/src/clt_elem.c"
#include "../../third_party/clt13/src/clt.c"

int go_clt_encode_level1(clt_elem_t *rop, const clt_state_t *state, unsigned long long value, const int *ix) {
    mpz_t values[1];
    mpz_init(values[0]);
    mpz_set_ui(values[0], value);
    int rc = clt_encode(rop, state, 1, (const mpz_t *) values, ix);
    mpz_clear(values[0]);
    return rc;
}

clt_state_t *go_clt_state_new(const clt_params_t *params, size_t flags) {
    aes_randstate_t rng;
    if (aes_randinit(rng) != AESRAND_OK) {
        return NULL;
    }
    clt_state_t *state = clt_state_new(params, NULL, 0, flags, rng);
    aes_randclear(rng);
    return state;
}

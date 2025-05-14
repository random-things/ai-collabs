#ifndef CSI_H
#define CSI_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct CSIHandle CSIHandle;

// Build and return a new handle (or NULL on error)
CSIHandle *csi_new(const uint8_t *data, size_t len);

// Free a handle
void csi_free(CSIHandle *h);

// Search: writes up to max_out offsets into out[], returns count
size_t csi_search(
    const CSIHandle *h,
    const uint8_t *pattern,
    size_t pat_len,
    size_t *out,
    size_t max_out
);

#ifdef __cplusplus
}
#endif

#endif // CSI_H
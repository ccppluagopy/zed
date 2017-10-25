int find(const char *src, const char *dst) {
	uint8_t keys[256] = {0};
	int dstlen = 0;
	uint8_t *psrc = (uint8_t*)src, *pdst = (uint8_t*)dst, tmp, mincount = 0xFFFFFFFF, pivotidx = 0;
	while(*psrc){
		keys[*psrc++]++;
	}

	while(tmp = *pdst++) {
		switch(keys[tmp]) {
		case 0:
			break;
		default:
			if (keys[tmp] > 0 && keys[tmp] < mincount) {
				mincount = keys[tmp];
				pivotidx = pdst - (uint8_t*)dst - 1;
			}
			break;
		}
	}

	dstlen = pdst - (uint8_t*)dst - 1;
	uint8_t key = dst[pivotidx];

	psrc = (uint8_t*)src;
	pdst = (uint8_t*)dst;

	if (pivotidx * 2 > dstlen) {
		while (tmp = *psrc) {
			if (tmp == key) {
				bool failed = false;
				for (int i = pivotidx, j = psrc - (uint8_t*)src; i < dstlen; i++, j++) {
					if (src[j] != dst[i]) {
						failed = true;
						break;
					}
				}
				if (!failed) {
					for (int i = pivotidx - 1, j = psrc - (uint8_t*)src - 1; i >= 0; i--, j--) {
						if (src[j] != dst[i]) {
							failed = true;
							break;
						}		
					}
				}
				if (!failed) {
					return psrc - (uint8_t*)src - pivotidx;
				}
			}
			psrc++;
		}
	} else {
		while (tmp = *psrc) {
			if (tmp == key) {
				bool failed = false;
				
				for (int i = pivotidx, j = psrc - (uint8_t*)src; i >= 0; i--, j--) {
					if (src[j] != dst[i]) {
						failed = true;
						break;
					}		
				}
				if (!failed) {
					for (int i = pivotidx + 1, j = psrc - (uint8_t*)src + 1; i < dstlen; i++, j++) {
						if (src[j] != dst[i]) {
							failed = true;
							break;
						}
					}
				}
				if (!failed) {
					return psrc - (uint8_t*)src - pivotidx;
				}
			}
			psrc++;
		}
	}
	return -1;
}
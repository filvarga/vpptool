/*
 * Copyright (c) 2020 Cisco and/or its affiliates.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdio.h>

#define SIZEOF_BIHASH_16_8_T 184
#define BITS(x) (8*sizeof(x))
#define count_leading_zeros(x) __builtin_clzll (x)

typedef unsigned int u32;

static inline u32
min_log2 (u32 x)
{
  u32 n;
  n = count_leading_zeros (x);
  return BITS (u32) - n - 1;
}

static inline u32
max_log2 (u32 x)
{
  u32 l = min_log2 (x);
  if (x > ((u32) 1 << l))
    l++;
  return l;
}

static u32
nat_calc_bihash_buckets (u32 n_elts)
{
  return 1 << (max_log2 (n_elts >> 1) + 1);
}

static u32
nat_calc_bihash_memory (u32 n_buckets, u32 kv_size)
{
  return n_buckets * (8 + kv_size * 4);
}

static u32
calc_memory_size (u32 translations)
{
  return nat_calc_bihash_memory (
      nat_calc_bihash_buckets (translations), SIZEOF_BIHASH_16_8_T);
}

#define DIM(x) (sizeof(x)/sizeof(*(x)))

static const char *sizes[] = {
  "EiB", "PiB", "TiB", "GiB", "MiB", "KiB", "B"
};
static const uint64_t exbibytes = 1024ULL * 1024ULL * 1024ULL *
                                  1024ULL * 1024ULL * 1024ULL;

char *
format_memory_size (uint64_t size)
{   
  // fix it
  char *result = (char *) malloc(sizeof(char) * 20);
  uint64_t multiplier = exbibytes;
  int i;

  for (i = 0; i < DIM(sizes); i++, multiplier /= 1024)
    {   
      if (size < multiplier)
        continue;
      if (size % multiplier == 0)
        sprintf(result, "%llu %s", size / multiplier, sizes[i]);
      else
        sprintf(result, "%.1f %s", (float) size / multiplier, sizes[i]);
      return result;
    }
  strcpy(result, "0");
  return result;
}

int
main (int argc, char ** argv)
{
  u32 l;
  char *s;

  if (argc < 2)
    exit (1);

  l = (u32) atoi (argv[1]);

  printf ("sessions: %u\n", l);

  l = calc_memory_size (l);
  s = format_memory_size (l);

  printf ("expected memory size: %s\n", s);

  free (s);
  exit (0);
}

import binascii
import csv
import random
from hashlib import sha256

SHUFFLE_ROUND_COUNT = 90


def bytes_to_int(data: bytes) -> int:
    return int.from_bytes(data, 'little')


def int_to_bytes1(x):
    return x.to_bytes(1, 'little')


def int_to_bytes4(x):
    return x.to_bytes(4, 'little')


def hash(data: bytes) -> bytes:
    return sha256(data).digest()


def get_permuted_index(index: int, list_size: int, seed: bytes) -> int:
    """
    Return `p(index)` in a pseudorandom permutation `p` of `0...list_size-1` with ``seed`` as entropy.

    Utilizes 'swap or not' shuffling found in
    https://link.springer.com/content/pdf/10.1007%2F978-3-642-32009-5_1.pdf
    See the 'generalized domain' algorithm on page 3.
    """
    assert index < list_size
    assert list_size <= 2 ** 40

    for round in range(SHUFFLE_ROUND_COUNT):
        pivot = bytes_to_int(hash(seed + int_to_bytes1(round))[0:8]) % list_size
        flip = (pivot - index) % list_size
        position = max(index, flip)
        source = hash(seed + int_to_bytes1(round) + int_to_bytes4(position // 256))
        byte = source[(position % 256) // 8]
        bit = (byte >> (position % 8)) % 2
        index = flip if bit else index

    return index


with open('tests.csv', mode='w') as employee_file:
    tests_writer = csv.writer(employee_file, delimiter=',', quotechar='"', quoting=csv.QUOTE_MINIMAL)

    for seed in [hash(int_to_bytes4(seed_init_value)) for seed_init_value in range(30)]:
        for list_size in [0, 1, 2, 3, 5, 10, 100, 1000]:
            start_list = [i for i in range(list_size)]
            # random input, using python shuffle. Seed is static here, we just want consistent test generation.
            # Checking the shuffling on a simple incremental list is not good enough.
            # I.e. we want the shuffle to be independent of the contents of the list.
            random.seed(123)
            random.shuffle(start_list)
            encoded_start = ":".join([str(x) for x in start_list])
            shuffling = [0 for _ in range(list_size)]
            for i in range(list_size):
                shuffling[get_permuted_index(i, list_size, seed)] = i
            end_list = [start_list[x] for x in shuffling]
            encoded_shuffled = ":".join([str(v) for v in end_list])
            tests_writer.writerow([binascii.hexlify(seed).decode("utf-8"), list_size, encoded_start, encoded_shuffled])

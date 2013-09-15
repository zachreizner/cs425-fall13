import sys
import string
import random


values = ["first line",
          "second line",
          "pizza clicker",
          "peppironi"]


def generate_logfile(filename, linelength, numlines):
    with open(filename, 'w') as f:
        for valstring in values:
            f.write(randomString(int(linelength/2)) + ':' + valstring + '\n')
        for i in range(numlines):
            f.write(gen_logline(linelength) + '\n')


def gen_logline(linelength):
    return randomString(int(linelength/2)) + ':' + ('a' * int(linelength/2))


def randomString(numchars):
    return ''.join(random.choice(string.ascii_uppercase + string.digits)
           for x in range(numchars))


def main(filenames):
    for name in filenames:
        generate_logfile(name, 10240, 10240)


if __name__ == "__main__":
    main(sys.argv[1:])

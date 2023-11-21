local opts = {
  Pool: {
    // Multiple memory pools with the same tag will be merged
    Tag: 'default',
    // Memory block size
    Cap: 1024 * 128,
    // How many blocks to cache
    Len: 8 * 10,
  },
};
[
    opts + {
        // connect portal address
        Portal: '127.0.0.1:4000',
        // 'portal' validates the token to determine if it is legitimate
        Token: 'any string',
    }
]
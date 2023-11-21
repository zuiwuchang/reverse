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
  {
    // listen addr, 'bridge' will connect this addr
    Addr: ':4000',
    // 'bridge' needs to pass in this token to verify that it is legitimate
    Token: 'any string',
    // The reused transmission channel will no longer transmit data for new connections after the specified number of seconds
    MaxSeconds: 60 * 10,
    // The reused transmission channel will no longer transmit data for new connections after specifying MB data.
    MaxMB: 1024 * 10,
    // Data forwarding target
    Forwards: [
      opts + {
        From: ':4104',
        To: 'tcp://127.0.0.1:80',
      },
    ],
  },
]

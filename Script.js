function main(config, profileName) {
  config.proxies = config.proxies || [];
  config['proxy-groups'] = config['proxy-groups'] || [];
  config.rules = config.rules || [];

  const ali_gost = {
    name: "HENU VPN",
    type: "ss",
    server: "",
    port: 10806,
    cipher: "chacha20-ietf-poly1305",
    password: "",
    udp: true
  };

  config.proxies = config.proxies.filter(p => p.name !== ali_gost.name);
  config.proxies.push(ali_gost);

  const henu_group = {
    name: "校园网",
    type: "select",
    url: "https://zszx.henu.edu.cn",
    proxies: ["DIRECT", ali_gost.name]
  };

  config['proxy-groups'].splice(1, 0, henu_group);

  config.rules.unshift(
    'DOMAIN,vpn.henu.edu.cn,DIRECT',
    'DOMAIN-SUFFIX,henu.edu.cn,校园网',
    'IP-CIDR,10.0.0.0/8,校园网,no-resolve',
  );

  return config;
}

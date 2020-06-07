import Mercury from '@postlight/mercury-parser';

import fastify from 'fastify'

const server = fastify()

server.post('/', async (req, reply) => {
  console.log("Analyzing", req.body.url)
  return {...await Mercury.parse(req.body.url), content: ''};
})

server.get('/', async (req, reply)=>{
  console.log('req', req)
})

server.listen(process.env.PORT || "8080", '0.0.0.0', (err, address) => {
  if(err) {
    console.error(err)
    process.exit(1)
  }
  console.log(`Server listening at ${address}`)
})

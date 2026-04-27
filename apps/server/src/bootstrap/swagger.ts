import { DocumentBuilder } from '@nestjs/swagger';

export const config = new DocumentBuilder()
  .setTitle('Lunie')
  .setDescription('The Lunie API description')
  .addTag('Lunie')
  .addBearerAuth(
    {
      type: 'http',
      scheme: 'bearer',
      bearerFormat: 'API Token',
      description: 'Paste your Lunie API token here',
    },
    'bearer', // <- name of the security scheme
  )
  .build();

export const SWAGGER_ENDPOINT = '/api';

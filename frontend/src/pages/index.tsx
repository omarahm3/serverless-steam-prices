import type { NextPage } from 'next';
import Head from 'next/head';
import { ChangeEvent, useState } from 'react';

type CardProps = {
  name: string;
  price: number;
};

const API_URL = process.env.NEXT_PUBLIC_API;
const IDLE_TIME = 1000;

interface App {
  appid: string;
  name: string;
  price: number;
}

interface Response {
  apps: App[];
  total: number;
}

const Home: NextPage = () => {
  const [tid, setTid] = useState<NodeJS.Timeout | undefined>(undefined);
  const [apps, setApps] = useState<Array<App>>([]);

  const request = async (query: string): Promise<Response> => {
    const response = await fetch(`${API_URL}?query=${query}`);
    return await response.json() as Response;
  };

  const search = (e: ChangeEvent<HTMLInputElement>) => {
    e.stopPropagation();
    const query = e.target.value;
    clearTimeout(tid);

    if (query.length <= 2) {
      return;
    }

    const id = setTimeout(async () => {
      console.log('Making the request with query:', query);
      const res = await request(query);
      if (res.total === 0) {
        return;
      }

      setApps(res.apps);
    }, IDLE_TIME);

    setTid(id);
  };

  return (
    <>
      <Head>
        <title>Steam Prices</title>
        <meta name="description" content="Generated by create-t3-app" />
        <link rel="icon" href="/favicon.ico" />
      </Head>

      <main className="container mx-auto flex flex-col items-center justify-center h-screen p-4">
        <h1 className="text-5xl md:text-[5rem] leading-normal font-extrabold text-gray-700">
          The unofficial <span className="text-blue-600">Steam</span> DB
        </h1>

        <div className="grid gap-3 pt-3 mt-3 text-center md:grid-cols-3 lg:w-2/3">
          <div className="col-span-3">
            <input
              type="text"
              placeholder="Apex Legends"
              className="mt-1 focus:ring-indigo-500 py-2 px-3 focus:border-indigo-500 block w-full shadow-sm sm:text-sm border border-gray-500 rounded-md"
              onChange={search}
            />
            <small>Make sure to enter more than 2 characters huh!</small>
          </div>
          {apps.length ? apps.map(app => <Card key={app.appid} name={app.name} price={app.price} />) : ''}
        </div>
      </main>
    </>
  );
};

const Card = ({ name, price }: CardProps) => {
  return (
    <section className="flex flex-col justify-center p-6 duration-500 border-2 border-gray-500 rounded shadow-xl motion-safe:hover:scale-105">
      <h2 className="text-lg text-gray-700">{name}</h2>
      <p className="text-sm text-gray-600">{price}</p>
    </section>
  );
};

export default Home;

// import { useNodeStore } from '@/store/nodeStore';
import Intl from '@/i18n/i18n';
import {SelectNetwork} from '@/components/SelectNetwork';
import OnOffBtn from '@/components/OnOffBtn';
import AutoRestart from '@/components/AutoRestart';

export default function Home() {
// const {status, network} = useNodeStore();


  return (
    <>
      {/* {isLoading ? (
        <Loading />
      ) : ( */}
        <div className="grid grid-cols-2 gap-5">
          <div className="bg-secondary rounded-2xl w-full max-w-lg p-10">
            <p className="mas-body text-f-primary mb-2">
              {Intl.t('home.title-select-network')}
            </p>
            <SelectNetwork />
          </div>

            <div className="bg-secondary rounded-2xl w-full max-w-lg p-10 flex flex-col justify-center">
            <OnOffBtn />
            </div>

          <div className="bg-secondary rounded-2xl w-full max-w-lg p-10 flex flex-col justify-center">
            <AutoRestart />
          </div>

          <div className="bg-secondary rounded-2xl w-full max-w-lg p-10">
            <p className="mas-body text-f-primary mb-2">
              {Intl.t('home.title-account-balance')}
            </p>
          </div>
        </div>
      {/* )} */}
    </>
  );
}

import {OffersListProps} from '../../types';
import styles from './PriceOffer.module.css';

const PriceOffer = ({offers}:OffersListProps)=>{
    return (
        <div className={styles.offersList}>
          {offers.map((offer, index) => (
            <div key={index} className={styles.offerItem}>
              <div className={styles.offerShop}>{offer.shopName}</div>
              <div className={styles.offerPrice}>
                {offer.price.toLocaleString()}
              </div>
              <div className={styles.offerAvailability}>{offer.availability}</div>
              <a 
                href={offer.url} 
                target="_blank" 
                rel="noopener noreferrer" 
                className={styles.offerLink}
              >
                В магазин
              </a>
            </div>
          ))}
        </div>
    );
}

export default PriceOffer;